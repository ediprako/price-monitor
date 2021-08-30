package pgsql

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

type ProductPayload struct {
	Name          string
	CurrentPrice  int64
	OriginalPrice int64
	Images        []string
}

type Product struct {
	ID            int64  `db:"id"`
	Name          string `db:"name"`
	CurrentPrice  int64  `db:"current_price"`
	OriginalPrice int64  `db:"original_price"`
	Images        []string
}

func (r *repository) GetProductsByID(ctx context.Context, id int64) (Product, error) {
	sql := `SELECT id, name, current_price, original_price FROM
		product WHERE id=$1`

	var product Product
	err := r.db.GetContext(ctx, &product, sql, id)
	if err != nil {
		return Product{}, err
	}

	sqlImage := `SELECT image FROM product_images WHERE product_id=$1`
	rows, err := r.db.QueryxContext(ctx, sqlImage, id)
	if err != nil {
		return Product{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var image string
		rows.Scan(&image)
		product.Images = append(product.Images, image)
	}

	return product, err
}

func (r *repository) GetProducts(ctx context.Context, limit, offset int) ([]Product, error) {
	sql := `SELECT id, name, current_price, original_price FROM
		product LIMIT $1 OFFSET $2`

	var products []Product
	err := r.db.SelectContext(ctx, &products, sql, limit, offset)
	if err != nil {
		return nil, err
	}

	listProductID := make([]int64, len(products))
	for i, product := range products {
		listProductID[i] = product.ID
	}

	sqlImage := `SELECT product_id,image FROM product_images WHERE product_id = ANY($1)`
	rows, err := r.db.QueryxContext(ctx, sqlImage, pq.Array(listProductID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mapProductImage := make(map[int64][]string)
	for rows.Next() {
		var productID int64
		var image string
		err = rows.Scan(&productID, &image)
		if err != nil {
			return nil, err
		}
		if mapProductImage[productID] == nil {
			mapProductImage[productID] = []string{}
		}
		mapProductImage[productID] = append(mapProductImage[productID], image)
	}

	for _, product := range products {
		product.Images = mapProductImage[product.ID]
	}

	return products, nil
}

func (r *repository) GetTotalProduct(ctx context.Context) (int64, error) {
	sql := `SELECT count(*) as total FROM
		product`

	var total int64
	err := r.db.QueryRowContext(ctx, sql).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *repository) UpsertProduct(ctx context.Context, payload ProductPayload) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	sql := `INSERT INTO product (name, current_price, original_price) VALUES 
		( $1, $2, $3) ON CONFLICT (name) DO UPDATE SET current_price = $2, original_price = $3
		RETURNING id`

	var id int64
	err = tx.QueryRowContext(ctx, sql, payload.Name, payload.CurrentPrice, payload.OriginalPrice).Scan(&id)
	if err != nil {
		return 0, err
	}

	sqlImages := `INSERT INTO product_images (product_id, image) VALUES ($1, $2)`
	for _, image := range payload.Images {
		_, err := tx.ExecContext(ctx, sqlImages, id, image)
		if err != nil {
			return 0, err
		}
	}

	sqlHistory := `INSERT INTO price_history(product_id, current_price, original_price) 
		VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, sqlHistory, id, payload.CurrentPrice, payload.OriginalPrice)
	if err != nil {
		return 0, err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) InsertPriceHistory(ctx context.Context, productID int64, currentPrice int64, originalPrice int64) error {
	sql := `INSERT INTO price_history(product_id, current_price, original_price) 
		VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, sql, productID, currentPrice, originalPrice)

	return err
}

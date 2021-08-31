package pgsql

import (
	"context"
	"database/sql"
	"time"

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
	URL           string
	Images        []string
}

type Product struct {
	ID            int64  `db:"id"`
	Name          string `db:"name"`
	CurrentPrice  int64  `db:"current_price"`
	OriginalPrice int64  `db:"original_price"`
	URL           string `db:"url"`
	Images        []string
}

type PriceHistory struct {
	ID            int64     `db:"id"`
	ProductID     int64     `db:"product_id"`
	CurrentPrice  int64     `db:"current_price"`
	OriginalPrice int64     `db:"original_price"`
	UpdateTime    time.Time `db:"updated_at"`
}

type ProductImage struct {
	ID        int64  `db:"id"`
	ProductID int64  `db:"product_id"`
	Image     string `db:"image"`
}

const (
	stateActive  = 1
	stateDeleted = 0
)

func (r *repository) GetProductsByID(ctx context.Context, id int64) (Product, error) {
	sql := `SELECT id, name, current_price, original_price,coalesce(url,'') url FROM
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

func (r *repository) GetProductsByUpdateTime(ctx context.Context, startTime, endTime time.Time) ([]Product, error) {
	sql := `SELECT id, name, current_price, original_price,coalesce(url,'') url FROM
		product WHERE updated_at between $1 AND $2`

	var product []Product
	err := r.db.SelectContext(ctx, &product, sql, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *repository) GetProducts(ctx context.Context, limit, offset int) ([]Product, error) {
	sql := `SELECT id, name, current_price, original_price, coalesce(url,'') url FROM
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

	sqlImage := `SELECT product_id,image FROM product_images WHERE product_id = ANY($1) and status=$2`
	rows, err := r.db.QueryxContext(ctx, sqlImage, pq.Array(listProductID), stateActive)
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

func (r *repository) GetImagesByProductID(ctx context.Context, productID int64) ([]ProductImage, error) {
	sql := `SELECT id, product_id, image FROM product_images WHERE product_id=$1`

	var images []ProductImage
	err := r.db.SelectContext(ctx, &images, sql, productID)
	if err != nil {
		return nil, err
	}

	return images, nil
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

	productID, err := r.InsertProduct(ctx, tx, payload)
	if err != nil {
		return 0, err
	}

	err = r.insertingImages(ctx, tx, payload, productID)
	if err != nil {
		return 0, err
	}

	sqlHistory := `INSERT INTO price_history(product_id, current_price, original_price) 
		VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, sqlHistory, productID, payload.CurrentPrice, payload.OriginalPrice)
	if err != nil {
		return 0, err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return 0, err
	}

	return productID, nil
}

func (r *repository) insertingImages(ctx context.Context, tx *sql.Tx, payload ProductPayload, productID int64) error {
	images, err := r.GetImagesByProductID(ctx, productID)
	if err != nil {
		return err
	}

	mapImages := make(map[string]int64)
	for _, image := range images {
		mapImages[image.Image] = image.ID
	}

	for _, image := range payload.Images {
		if _, ok := mapImages[image]; ok {
			delete(mapImages, image)
			continue
		}

		err = r.AddNewProductImage(ctx, tx, productID, image)
		if err != nil {
			return err
		}
	}

	for _, val := range mapImages {
		err = r.SoftDeleteProductImage(ctx, tx, val)
		if err != nil {
			return err
		}
	}
	return err
}

func (r *repository) AddNewProductImage(ctx context.Context, tx *sql.Tx, productID int64, image string) error {
	sqlImages := `INSERT INTO product_images (product_id, image, status) VALUES ($1, $2 , $3)`
	_, err := tx.ExecContext(ctx, sqlImages, productID, image, stateActive)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) InsertProduct(ctx context.Context, tx *sql.Tx, payload ProductPayload) (int64, error) {
	sql := `INSERT INTO product (name, current_price, original_price,url) VALUES 
		( $1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET current_price = $2, original_price = $3, url = $4
		RETURNING id`

	var id int64
	err := tx.QueryRowContext(ctx, sql, payload.Name, payload.CurrentPrice, payload.OriginalPrice, payload.URL).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repository) SoftDeleteProductImage(ctx context.Context, tx *sql.Tx, id int64) error {
	sqlImages := `UPDATE product_images SET status = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, sqlImages, stateDeleted, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) InsertPriceHistory(ctx context.Context, productID int64, currentPrice int64, originalPrice int64) error {
	sql := `INSERT INTO price_history(product_id, current_price, original_price) 
		VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, sql, productID, currentPrice, originalPrice)

	return err
}

func (r *repository) GetLastPriceHistory(ctx context.Context, productID int64, limit int) ([]PriceHistory, error) {
	sql := `SELECT id, product_id, current_price, original_price, updated_at FROM (SELECT id, product_id, current_price, original_price, updated_at FROM
		price_history WHERE product_id = $1 
		ORDER BY updated_at DESC 
		LIMIT $2) p ORDER BY updated_at ASC`
	var histories []PriceHistory
	err := r.db.SelectContext(ctx, &histories, sql, productID, limit)
	if err != nil {
		return nil, err
	}

	return histories, nil
}

func (r *repository) UpDatabase(ctx context.Context) error {
	sql := `CREATE TABLE IF NOT EXISTS public.product (
		id bigserial NOT NULL,
		"name" varchar NOT NULL,
		current_price int8 NOT NULL,
		original_price int8 NOT NULL,
		updated_at information_schema."time_stamp" NOT NULL DEFAULT now(),
		url varchar NULL,
		CONSTRAINT product_pk PRIMARY KEY (id),
		CONSTRAINT product_un UNIQUE (name)
	)`
	_, err := r.db.ExecContext(ctx, sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS public.price_history (
		id bigserial NOT NULL,
		product_id int8 NOT NULL,
		current_price int8 NOT NULL,
		original_price int8 NOT NULL,
		updated_at information_schema."time_stamp" NOT NULL DEFAULT now()
	)`

	_, err = r.db.ExecContext(ctx, sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS public.product_images (
		id bigserial NOT NULL,
		product_id int8 NOT NULL,
		image varchar NOT NULL,
		status int NOT NULL,
		CONSTRAINT product_images_pk PRIMARY KEY (id)
	)`

	_, err = r.db.ExecContext(ctx, sql)
	if err != nil {
		return err
	}

	return nil
}

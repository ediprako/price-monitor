package usecase

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/ediprako/pricemonitor/repository/pgsql"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type dbProvider interface {
	GetProductsByID(ctx context.Context, id int64) (pgsql.Product, error)
	GetProducts(ctx context.Context, limit, offset int) ([]pgsql.Product, error)
	UpsertProduct(ctx context.Context, payload pgsql.ProductPayload) (int64, error)
	InsertPriceHistory(ctx context.Context, productID int64, currentPrice int64, originalPrice int64) error
	GetTotalProduct(ctx context.Context) (int64, error)
}

type usecase struct {
	db dbProvider
}

func New(db dbProvider) *usecase {
	return &usecase{
		db: db,
	}
}

type ProductPayload pgsql.ProductPayload
type Product struct {
	ID                  int64    `json:"id"`
	Name                string   `json:"name"`
	CurrentPrice        int64    `json:"current_price"`
	CurrentPriceString  string   `json:"current_price_string"`
	OriginalPrice       int64    `json:"original_price"`
	OriginalPriceString string   `json:"original_price_string"`
	Images              []string `json:"images,omitempty"`
}

func (u *usecase) RegisterProduct(ctx context.Context, link string) error {
	err, product, err2 := u.getProductFromLink(link)
	if err2 != nil {
		return err2
	}

	_, err = u.db.UpsertProduct(ctx, pgsql.ProductPayload(product))
	if err != nil {
		return err
	}
	return nil
}

func (u *usecase) GetProductDetail(ctx context.Context, id int64) (Product, error) {
	product, err := u.db.GetProductsByID(ctx, id)
	if err != nil {
		return Product{}, err
	}

	result := Product{
		ID:            product.ID,
		Name:          product.Name,
		CurrentPrice:  product.CurrentPrice,
		OriginalPrice: product.OriginalPrice,
		Images:        product.Images,
	}

	return result, nil
}

type PaginateData struct {
	Draw            string    `json:"draw"`
	RecordsTotal    int64     `json:"recordsTotal"`
	RecordsFiltered int64     `json:"recordsFiltered"`
	Products        []Product `json:"data"`
}

func (u *usecase) ListProduct(ctx context.Context, draw string, page, pagesize int) (PaginateData, error) {
	if page == 0 {
		page = 1
	}
	offset := (page - 1) * pagesize
	products, err := u.db.GetProducts(ctx, pagesize, offset)
	if err != nil {
		return PaginateData{}, err
	}
	total, err := u.db.GetTotalProduct(ctx)

	if err != nil {
		return PaginateData{}, err
	}

	var paging PaginateData
	paging.RecordsTotal = total
	paging.RecordsFiltered = total
	paging.Draw = draw

	var result []Product
	for _, product := range products {
		result = append(result, Product{
			ID:            product.ID,
			Name:          product.Name,
			CurrentPrice:  product.CurrentPrice,
			OriginalPrice: product.OriginalPrice,
		})
	}

	paging.Products = result

	return paging, nil
}

func (u *usecase) getProductFromLink(link string) (error, ProductPayload, error) {
	response, err := http.Get(link)
	if err != nil {
		return nil, ProductPayload{}, err
	}
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, ProductPayload{}, err
	}

	var product ProductPayload

	product.Name = doc.Find("h1#product-name").Text()
	product.CurrentPrice = convertToAngka(doc.Find("div#product-final-price").First().Text())
	product.OriginalPrice = convertToAngka(doc.Find("div#product-discount-price").First().Text())
	doc.Find(".css-1iv32ek").Children().Each(func(i int, sel *goquery.Selection) {
		srcCrop, _ := sel.Find("img#product-image").Attr("src")
		sliceSrc := strings.Split(srcCrop, "&")
		if len(sliceSrc) > 0 {
			product.Images = append(product.Images, sliceSrc[0])
		}
	})
	return err, product, nil
}

func convertToAngka(rupiah string) int64 {
	m1 := regexp.MustCompile(`,.*|\D`)
	str := m1.ReplaceAllString(rupiah, "")
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

package usecase

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/ediprako/pricemonitor/repository/pgsql"
)

type dbProvider interface {
	GetProductsByID(ctx context.Context, id int64) (pgsql.Product, error)
	GetProductsByUpdateTime(ctx context.Context, startTime, endTime time.Time) ([]pgsql.Product, error)
	GetProducts(ctx context.Context, limit, offset int) ([]pgsql.Product, error)
	UpsertProduct(ctx context.Context, payload pgsql.ProductPayload) (int64, error)
	InsertPriceHistory(ctx context.Context, productID int64, currentPrice int64, originalPrice int64) error
	GetTotalProduct(ctx context.Context) (int64, error)
	GetLastPriceHistory(ctx context.Context, productID int64, limit int) ([]pgsql.PriceHistory, error)
	UpDatabase(ctx context.Context) error
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
	URL                 string   `json:"url"`
	Images              []string `json:"images,omitempty"`
}

type PaginateData struct {
	Draw            string    `json:"draw"`
	RecordsTotal    int64     `json:"recordsTotal"`
	RecordsFiltered int64     `json:"recordsFiltered"`
	Products        []Product `json:"data"`
}

type PriceHistory struct {
	ID            int64  `json:"id"`
	ProductID     int64  `json:"product_id"`
	CurrentPrice  int64  `json:"current_price"`
	OriginalPrice int64  `json:"original_price"`
	UpdateTime    string `json:"update_time"`
}

func (u *usecase) RegisterProduct(ctx context.Context, link string) (int64, error) {
	product, err := u.getProductFromLink(link)
	if err != nil {
		return 0, err
	}

	id, err := u.db.UpsertProduct(ctx, pgsql.ProductPayload(product))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (u *usecase) GetProductDetail(ctx context.Context, id int64) (Product, error) {
	product, err := u.db.GetProductsByID(ctx, id)
	if err != nil {
		return Product{}, err
	}

	result := Product{
		ID:                  product.ID,
		Name:                product.Name,
		CurrentPrice:        product.CurrentPrice,
		OriginalPrice:       product.OriginalPrice,
		Images:              product.Images,
		URL:                 product.URL,
		OriginalPriceString: "Rp. " + humanize.Comma(product.OriginalPrice),
		CurrentPriceString:  "Rp. " + humanize.Comma(product.CurrentPrice),
	}

	return result, nil
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
			URL:           product.URL,
		})
	}

	paging.Products = result

	return paging, nil
}

func (u *usecase) getProductFromLink(link string) (ProductPayload, error) {
	response, err := http.Get(link)
	if err != nil {
		return ProductPayload{}, err
	}
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return ProductPayload{}, err
	}

	var product ProductPayload

	product.Name = doc.Find("h1#product-name").Text()
	product.CurrentPrice = convertToAngka(doc.Find("div#product-final-price").First().Text())
	product.OriginalPrice = convertToAngka(doc.Find("div#product-discount-price").First().Text())
	if product.OriginalPrice == 0 {
		product.OriginalPrice = product.CurrentPrice
	}

	// find images
	doc.Find(".css-1iv32ek").Children().Each(func(i int, sel *goquery.Selection) {
		srcCrop, _ := sel.Find("img#product-image").Attr("src")
		sliceSrc := strings.Split(srcCrop, "&")
		if len(sliceSrc) > 0 {
			product.Images = append(product.Images, sliceSrc[0])
		}
	})

	product.URL = link
	return product, err
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

func (u *usecase) ListPriceHistory(ctx context.Context, productID int64, limit int) ([]PriceHistory, error) {
	if limit == 0 {
		limit = 100
	}
	histories, err := u.db.GetLastPriceHistory(ctx, productID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]PriceHistory, len(histories))
	for i, history := range histories {
		result[i] = PriceHistory{
			ID:            history.ID,
			ProductID:     history.ProductID,
			CurrentPrice:  history.CurrentPrice,
			OriginalPrice: history.OriginalPrice,
			UpdateTime:    history.UpdateTime.Format("2006-01-02 15:04"),
		}
	}

	return result, nil
}

func (u *usecase) UpDatabase(ctx context.Context) error {
	return u.db.UpDatabase(ctx)
}

func (u *usecase) RefreshProductInformation(ctx context.Context) error {
	startDate := time.Now().Add(time.Duration(-1) * time.Hour)
	startDate = startDate.Add(time.Duration(startDate.Second()*-1) * time.Second)
	endDate := startDate.Add(time.Duration(1) * time.Minute)

	products, err := u.db.GetProductsByUpdateTime(ctx, startDate, endDate)
	if err != nil {
		return err
	}

	if len(products) == 0 {
		log.Println("no product updated")
		return nil
	}

	for _, product := range products {
		_, err = u.RegisterProduct(ctx, product.URL)
		if err != nil {
			return err
		}
	}

	return nil
}

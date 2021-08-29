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
	m1 := regexp.MustCompile(`/,.*|[^0-9]/g`)
	str := m1.ReplaceAllString(rupiah, "")
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

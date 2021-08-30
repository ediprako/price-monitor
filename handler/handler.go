package handler

import (
	"context"
	"html/template"
	"net/http"
	"path"
	"strconv"

	httpHandler "github.com/ediprako/pricemonitor/handler/http"
	"github.com/ediprako/pricemonitor/usecase"
)

type usecaseProvider interface {
	RegisterProduct(ctx context.Context, link string) error
	ListProduct(ctx context.Context, draw string, page, pagesize int) (usecase.PaginateData, error)
	GetProductDetail(ctx context.Context, id int64) (usecase.Product, error)
	ListPriceHistory(ctx context.Context, productID int64, limit int) ([]usecase.PriceHistory, error)
}

type handler struct {
	usecase usecaseProvider
}

func New(usecase usecaseProvider) *handler {
	return &handler{
		usecase: usecase,
	}
}

func (h *handler) HandleIndexView(w http.ResponseWriter, _ *http.Request) {
	var tmpl = template.Must(template.ParseFiles(
		path.Join("handler", "ui", "index.html"),
		path.Join("handler", "ui", "navbar.html"),
	))

	var data = map[string]interface{}{}

	err := tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func (h *handler) HandleListView(w http.ResponseWriter, _ *http.Request) {
	var tmpl = template.Must(template.ParseFiles(
		path.Join("handler", "ui", "list.html"),
		path.Join("handler", "ui", "navbar.html"),
	))

	var data = map[string]interface{}{}

	err := tmpl.ExecuteTemplate(w, "list", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (h *handler) HandleDetailView(w http.ResponseWriter, r *http.Request) {
	var tmpl = template.Must(template.ParseFiles(
		path.Join("handler", "ui", "detail.html"),
		path.Join("handler", "ui", "navbar.html"),
	))

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	product, err := h.usecase.GetProductDetail(r.Context(), id)
	var data = map[string]interface{}{
		"product": product,
	}

	err = tmpl.ExecuteTemplate(w, "detail", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (h *handler) HandleAddLink(w http.ResponseWriter, r *http.Request) {
	inputLink := r.FormValue("input_link")
	err := h.usecase.RegisterProduct(r.Context(), inputLink)
	if err != nil {
		httpHandler.WriteHTTPResponse(w, nil, err, http.StatusInternalServerError)
		return
	}

	httpHandler.WriteHTTPResponse(w, struct {
		ID int64
	}{1}, nil, http.StatusOK)
	return
}

func (h *handler) HandleListProduct(w http.ResponseWriter, r *http.Request) {
	start, _ := strconv.Atoi(r.FormValue("start"))
	length, _ := strconv.Atoi(r.FormValue("length"))
	draw := r.FormValue("draw")
	paginated, err := h.usecase.ListProduct(r.Context(), draw, start, length)
	if err != nil {
		httpHandler.WriteHTTPResponse(w, nil, err, http.StatusInternalServerError)
		return
	}

	httpHandler.WriteHTTPAjax(w, paginated, http.StatusOK)
}

func (h *handler) HandleListHistories(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.FormValue("product_id"), 10, 64)
	if err != nil {
		httpHandler.WriteHTTPResponse(w, nil, err, http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.FormValue("limit"))

	histories, err := h.usecase.ListPriceHistory(r.Context(), productID, limit)
	if err != nil {
		httpHandler.WriteHTTPResponse(w, nil, err, http.StatusInternalServerError)
		return
	}

	httpHandler.WriteHTTPAjax(w, histories, http.StatusOK)
}

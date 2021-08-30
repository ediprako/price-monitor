package handler

import (
	"context"
	httpHandler "github.com/ediprako/pricemonitor/handler/http"
	"github.com/ediprako/pricemonitor/usecase"
	"html/template"
	"net/http"
	"path"
	"strconv"
)

type usecaseProvider interface {
	RegisterProduct(ctx context.Context, link string) error
	ListProduct(ctx context.Context, draw string, page, pagesize int) (usecase.PaginateData, error)
	GetProductDetail(ctx context.Context, id int64) (usecase.Product, error)
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
	var filepath = path.Join("handler", "ui", "index.html")
	var tmpl, err = template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data = map[string]interface{}{}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func (h *handler) HandleListView(w http.ResponseWriter, _ *http.Request) {
	var filepath = path.Join("handler", "ui", "list.html")
	var tmpl, err = template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data = map[string]interface{}{}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (h *handler) HandleDetailView(w http.ResponseWriter, r *http.Request) {
	var filepath = path.Join("handler", "ui", "detail.html")
	var tmpl, err = template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	product, err := h.usecase.GetProductDetail(r.Context(), id)
	var data = map[string]interface{}{
		"product": product,
	}

	err = tmpl.Execute(w, data)
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
	return
}

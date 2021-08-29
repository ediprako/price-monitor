package handler

import (
	"context"
	httpHandler "github.com/ediprako/pricemonitor/handler/http"
	"html/template"
	"net/http"
	"path"
)

type usecaseProvider interface {
	RegisterProduct(ctx context.Context, link string) error
}

type handler struct {
	usecase usecaseProvider
}

func New(usecase usecaseProvider) *handler {
	return &handler{
		usecase: usecase,
	}
}

func (h *handler) HandleIndex(w http.ResponseWriter, _ *http.Request) {
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
	}
}

type Product struct {
	Name          string
	CurrentPrice  string
	OriginalPrice string
	Images        []string
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
}

package handler

import (
	"html/template"
	"net/http"
	"path"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	var filepath = path.Join("handler", "ui", "index.html")
	var tmpl, err = template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data = map[string]interface{}{
		"title": "Learning Golang Web",
		"name":  "Batman",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

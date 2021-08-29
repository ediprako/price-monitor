package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ediprako/pricemonitor/handler"
	"github.com/ediprako/pricemonitor/repository/pgsql"
	"github.com/ediprako/pricemonitor/usecase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func init() {
	fmt.Println("INIT")
}

func main() {
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("handler/assets"))))

	http.Handle("/images",
		http.StripPrefix("/images/",
			http.FileServer(http.Dir("handler/img"))))

	db, err := sqlx.Connect("postgres", "user=postgres password=root dbname=price_monitor sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	repoDB := pgsql.New(db)
	uc := usecase.New(repoDB)
	h := handler.New(uc)

	http.HandleFunc("/", h.HandleIndex)
	http.HandleFunc("/addlink", h.HandleAddLink)

	fmt.Println("server started at localhost:9000")
	http.ListenAndServe(":9000", nil)
}

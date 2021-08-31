package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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

	db, err := sqlx.Connect("postgres", "postgres://uaqvhewtenpctm:6323e54cc2b84b73ad2ebac5820e172c15624660d4990ce97031eeea462bd75a@ec2-54-156-60-12.compute-1.amazonaws.com:5432/dgk2vue5v330d")
	if err != nil {
		log.Fatalln(err)
	}

	repoDB := pgsql.New(db)
	uc := usecase.New(repoDB)
	h := handler.New(uc)

	err = h.HandleUpDatabase(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", h.HandleIndexView)
	http.HandleFunc("/listview", h.HandleListView)
	http.HandleFunc("/list/product", h.HandleListProduct)
	http.HandleFunc("/addlink", h.HandleAddLink)
	http.HandleFunc("/detailview", h.HandleDetailView)
	http.HandleFunc("/histories", h.HandleListHistories)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8999" // Default port if not specified
	}
	fmt.Println("server started at localhost: ", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

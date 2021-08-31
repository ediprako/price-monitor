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
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	fmt.Println("INIT")
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	user := os.Getenv("DBUSER")
	password := os.Getenv("DBPASSWORD")
	dbname := os.Getenv("DBNAME")
	host := os.Getenv("DBHOST")
	dbport := os.Getenv("DBPORT")
	sslmode := os.Getenv("DBSSLMODE")

	db, err := settingDB(user, password, dbname, host, dbport, sslmode)
	if err != nil {
		log.Fatal(err)
	}

	repoDB := pgsql.New(db)
	uc := usecase.New(repoDB)
	h := handler.New(uc)

	err = h.HandleUpDatabase(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("handler/assets"))))

	http.Handle("/images",
		http.StripPrefix("/images/",
			http.FileServer(http.Dir("handler/img"))))

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

func settingDB(user, password, dbname, host, port, ssl string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres",
		fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, ssl))
	if err != nil {
		return nil, err
	}

	return db, nil
}

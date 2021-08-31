package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ediprako/pricemonitor/handler"
	"github.com/ediprako/pricemonitor/handler/cron"
	"github.com/ediprako/pricemonitor/repository/pgsql"
	"github.com/ediprako/pricemonitor/usecase"
	"github.com/gorilla/mux"
	"github.com/jasonlvhit/gocron"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	fmt.Println("INIT")
}

func main() {
	mode := flag.String("mode", "http", "service mode (http,cron)")
	flag.Parse()

	if *mode == "" {
		*mode = "http"
	}

	switch *mode {
	case "http":
		mainHttp()
	case "cron":
		mainCron()
	default:
		log.Fatal("unknown mode")
	}
}

func mainCron() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	dbport := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	db, err := settingDB(user, password, dbname, host, dbport, sslmode)
	if err != nil {
		log.Fatal(err)
	}

	repoDB := pgsql.New(db)
	uc := usecase.New(repoDB)
	c := cron.New(uc)
	gocron.Every(1).Minutes().Do(c.CronRefreshProductInformation)

	<-gocron.Start()
}

func mainHttp() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	dbport := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

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

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("handler/assets"))))

	r.HandleFunc("/", h.HandleIndexView).Methods(http.MethodGet)
	r.HandleFunc("/listview", h.HandleListView).Methods(http.MethodGet)
	r.HandleFunc("/list/product", h.HandleListProduct).Methods(http.MethodGet)
	r.HandleFunc("/addlink", h.HandleAddLink).Methods(http.MethodPost)
	r.HandleFunc("/detailview", h.HandleDetailView).Methods(http.MethodGet)
	r.HandleFunc("/histories", h.HandleListHistories).Methods(http.MethodGet)
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
		return
	}).Methods(http.MethodGet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}
	fmt.Println("server started at localhost: ", port)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func settingDB(user, password, dbname, host, port, ssl string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres",
		fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, ssl))
	if err != nil {
		return nil, err
	}

	return db, nil
}

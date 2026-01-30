package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"algosphera/scanner-api/internal/db"
	httpapi "algosphera/scanner-api/internal/http"

	tickersets "algosphera/scanner-api/internal/tickersets"

	dotenv "github.com/joho/godotenv"
)

func main() {
	err := dotenv.Load()
	if err != nil {
		log.Printf("failed to load .env file: %v, continuing without .env", err)
	}

	catalogDir := os.Getenv("TICKER_SETS_DIR")
	if catalogDir == "" {
		catalogDir = "ticker_sets" // local dev default
	}
	cat, err := tickersets.LoadDir(catalogDir)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer pool.Close()

	router, err := httpapi.NewRouter(pool, cat)
	if err != nil {
		log.Fatalf("failed to init router: %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("scanner-api listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

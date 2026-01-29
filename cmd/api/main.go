package main

import (
	"log"
	"net/http"
	"os"
	"time"

	httpapi "algosphera/scanner-api/internal/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router, err := httpapi.NewRouter()
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

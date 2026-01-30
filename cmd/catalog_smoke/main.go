package main

import (
	"fmt"
	"log"

	tickersets "algosphera/scanner-api/internal/tickersets"
)

func main() {
	cat, err := tickersets.LoadDir("./ticker_sets")
	if err != nil {
		log.Fatal(err)
	}

	title, tickers, err := cat.Resolve("ironcondor_friendly.all")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(title)
	fmt.Println(tickers)
}

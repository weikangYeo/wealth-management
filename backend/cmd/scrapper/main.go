package main

import (
	"log"
	"wealth-management/internal/gold"
)

func main() {
	log.Println("Starting scrapper")
	gold.ScrapeGoldPrice()
}

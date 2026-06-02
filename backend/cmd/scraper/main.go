package main

import (
	"log"
	"wealth-management/internal/gold"
	"wealth-management/internal/platform/config"
	"wealth-management/internal/platform/database"
	"wealth-management/internal/stock"
)

func main() {
	db, err := database.InitDbConnection(false)
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}
	log.Println("Starting scraper")
	config.BootstrapCommonConfig()
	gold.ScrapeGoldPrice()
	stock.ScrapeStockLastDonePrice(db)
}

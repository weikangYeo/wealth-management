package main

import (
	"log"
	"wealth-management/internal/fund"
	"wealth-management/internal/gold"
	"wealth-management/internal/platform/config"
	"wealth-management/internal/platform/database"
	"wealth-management/internal/stock"
)

func main() {
	config.BootstrapCommonConfig()
	db, err := database.InitDbConnection(false)
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}
	log.Println("Starting scraper")
	gold.ScrapeGoldPrice()
	log.Println("Gold Price Scraped")
	stock.ScrapeStockLastDonePrice(db)
	log.Println("Stock Price and Income Dist Scraped")
	fund.ScrapeFundNavAndIncomeDist(db)
	log.Println("Fund Price and Income Dist Scraped")
}

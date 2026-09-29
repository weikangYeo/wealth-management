package main

import (
	"log"
	"wealth-management/internal/fund"
	"wealth-management/internal/platform/config"
	"wealth-management/internal/platform/database"
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
	//gold.ScrapeGoldPrice()
	//stock.ScrapeStockLastDonePrice(db)
	fund.ScrapeFundNavAndIncomeDist(db)
}

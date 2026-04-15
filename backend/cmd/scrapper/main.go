package main

import (
	"log"
	"wealth-management/internal/gold"
	"wealth-management/internal/platform/config"
	"wealth-management/internal/stock"
)

func main() {
	log.Println("Starting scrapper")
	config.BootstrapCommonConfig()
	gold.ScrapeGoldPrice()
	stock.ScrapeStockLastDonePrice()
}

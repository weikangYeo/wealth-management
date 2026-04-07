package main

import (
	"flag"
	"log"
	"wealth-management/internal/gold"
	"wealth-management/internal/platform/config"
)

func main() {
	isLive := flag.Bool("live", false, "live run, scrape from web page instead of mock")
	flag.Parse()
	log.Println("Starting scrapper")
	config.BootstrapCommonConfig()
	gold.ScrapeGoldPrice(*isLive)
}

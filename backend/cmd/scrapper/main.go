package main

import (
	"flag"
	"log"
	"wealth-management/internal/gold"
	"wealth-management/internal/platform/config"
)

func main() {
	isTrial := flag.Bool("trial", true, "trial run, scrape from static test page")
	flag.Parse()
	log.Println("Starting scrapper")
	config.BootstrapCommonConfig()
	gold.ScrapeGoldPrice(*isTrial)
}

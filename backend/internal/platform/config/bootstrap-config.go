package config

import (
	"log"

	"github.com/joho/godotenv"
)

func BootstrapCommonConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// set logger properties
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

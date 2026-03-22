package main

import (
	"log"
	"wealth-management/internal/app"
	"wealth-management/internal/platform/config"
	"wealth-management/internal/platform/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	startApp()
}

func startApp() {
	config.BootstrapCommonConfig()
	db, err := database.InitDbConnection(true)
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}

	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()
	//r.Use(cors.Default()) // All origins allowed by default
	corsConfigs := cors.DefaultConfig()
	corsConfigs.AllowOrigins = []string{"http://localhost:4200"}
	r.Use(cors.New(corsConfigs))

	// Point to main route
	app.SetupRoutes(r, db)
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

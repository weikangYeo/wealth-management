package main

import (
	"database/sql"
	"log"
	"os"
	"wealth-management/routes"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	// load property to env variable
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initDb()
	if db != nil {
		defer db.Close()
	}

	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Point to main route
	routes.SetupRoutes(r, db)
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func initDb() {
	// establish db connection and run db migration scripts
	// Capture connection properties.
	cfg := mysqlDriver.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "localhost:3307"
	cfg.DBName = "wealth_management"
	cfg.Params = map[string]string{}
	// so DB time (uint) value can be parsed to golang time
	cfg.Params["parseTime"] = "true"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("failed to open db connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("failed to create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../devops/database/migrations",
		"mysql",
		driver)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("Migrations ran successfully")
}

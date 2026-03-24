package database

import (
	"database/sql"
	"errors"
	"log"
	"os"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
)

func InitDbConnection(isRunMigration bool) (*sql.DB, error) {
	var db *sql.DB

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
	// when run migration multiple ddl can be executed in a same file
	cfg.MultiStatements = true

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return db, err
	}

	if err := db.Ping(); err != nil {
		return db, err
	}

	if isRunMigration {
		driver, err := mysql.WithInstance(db, &mysql.Config{})
		if err != nil {
			return db, err
		}

		m, err := migrate.NewWithDatabaseInstance(
			"file://../devops/database/migrations",
			"mysql",
			driver)
		if err != nil {
			return db, err
		}

		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return db, err
		}
		log.Println("Migrations ran successfully")
	}

	return db, err
}

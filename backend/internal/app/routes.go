package app

import (
	"database/sql"
	"wealth-management/internal/admin"
	"wealth-management/internal/gold"
	"wealth-management/internal/stock"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB) {
	admin.RegisterPingRoute(r)
	gold.RegisterGoldRoutes(r, db)
	stock.RegisterStockRoutes(r, db)
}

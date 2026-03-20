package routes

import (
	"database/sql"
	"wealth-management/internal/gold"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB) {
	registerPingRoute(r)
	gold.RegisterGoldRoutes(r, db)
}

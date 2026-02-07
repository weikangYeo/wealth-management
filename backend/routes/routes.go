package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB) {
	registerPingRoute(r)
	registerGoldRoutes(r, db)
}

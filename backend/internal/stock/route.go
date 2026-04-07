package stock

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterStockRoutes(r *gin.Engine, db *sql.DB) {
	repo := newStockRepository(db)
	stockHandler := newStockHandler(repo)
	stockRoutes := r.Group("/stocks")
	{
		stockRoutes.GET("", stockHandler.getAllStockSummary)
		stockRoutes.POST("", stockHandler.postStock)
	}
}

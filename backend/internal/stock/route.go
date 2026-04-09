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
		stockRoutes.GET("", stockHandler.getAllStock)
		stockRoutes.POST("", stockHandler.createStock)
		stockRoutes.GET("/:stockCode/transactions", stockHandler.getAllStockTransactions)
		stockRoutes.POST("/:stockCode/transactions", stockHandler.createStockTxn)
	}
}

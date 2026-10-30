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
		stockRoutes.GET("/:stockName/transactions", stockHandler.getAllStockTransactions)
		// rename to /:stockName to make it consistent with GET stocks which provide aggregate info
		stockRoutes.GET("/:stockName", stockHandler.getStock)
		stockRoutes.POST("/:stockName/transactions", stockHandler.createStockTxn)
		stockRoutes.PUT("/:stockName/transactions/:txnId", stockHandler.updateStockTxn)
		stockRoutes.GET("/:stockName/dividends", stockHandler.getDividendsByStockName)
		stockRoutes.POST("/:stockName/dividends", stockHandler.createStockDividend)
		stockRoutes.PUT("/:stockName/dividends", stockHandler.updateStockDividend)
	}
}

package stock

import (
	"log"
	"net/http"

	"github.com/cockroachdb/apd/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handler struct {
	stockRepo *repository
}

func newStockHandler(stockRepo *repository) *handler {
	return &handler{stockRepo: stockRepo}
}

// do a basic stock info listing for now
// todo enhance to show some aggregated info, probably we need to read from a pre-compute table for performance
func (handler *handler) getAllStock(context *gin.Context) {
	stocks, err := handler.stockRepo.getAllStocks()
	if err != nil {
		log.Printf("Error getting all stocks: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Retrieved %d stocks from database", len(stocks))
	for i, stock := range stocks {
		log.Printf("Stock %d: StockCode='%s', DisplayName='%s'", i, stock.StockCode, stock.DisplayName)
	}
	context.JSON(http.StatusOK, gin.H{"content": stocks})
}

// todo enhance to include more aggregated details
func (handler *handler) getStockOverview(context *gin.Context) {
	stock, err := handler.stockRepo.getStockByCode(context.Param("stockCode"))
	if err != nil {
		log.Printf("Error getting stock: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, stock)
}

func (handler *handler) getAllStockTransactions(context *gin.Context) {
	stockCode := context.Param("stockCode")
	stocks, err := handler.stockRepo.getStockTxnByStockCode(stockCode)
	if err != nil {
		log.Printf("Error getting all stocks summary: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"content": stocks})
}

func (handler *handler) createStock(context *gin.Context) {
	var req Stock
	if err := context.ShouldBindJSON(&req); err != nil {
		// consider change to context.AbortXxx
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := handler.stockRepo.createStock(req); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "created"})
}

func (handler *handler) createStockTxn(context *gin.Context) {
	stockCode := context.Param("stockCode")
	var req TxnRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn := Txn{
		ID:        uuid.New().String(),
		StockCode: stockCode,
		TxnDate:   req.TxnDate,
		Unit:      req.Unit,
		UnitPrice: req.UnitPrice,
		BrokerFee: req.BrokerFee,
		TxnType:   req.TxnType,
		Remark:    req.Remark,
	}

	// Calculate total price
	decimalCtx := apd.BaseContext.WithPrecision(12)
	if err := txn.CalculateTotalPrice(decimalCtx); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calculation: " + err.Error()})
		return
	}

	if err := handler.stockRepo.createStockTxn(txn); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "created"})
}

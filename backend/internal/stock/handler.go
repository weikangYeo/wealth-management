package stock

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	stockRepo *repository
}

func newStockHandler(stockRepo *repository) *handler {
	return &handler{stockRepo: stockRepo}
}

// do a basic stock info listing for now
// todo enhance to show some aggregated info, probably we need to read from a pre-compute table for performance
func (handler *handler) getAllStockSummary(context *gin.Context) {
	stocks, err := handler.stockRepo.getAllStockTxn()
	if err != nil {
		log.Printf("Error getting all stocks summary: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"content": stocks})
}

func (handler *handler) postStock(context *gin.Context) {
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

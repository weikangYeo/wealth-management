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

func (handler *handler) getAllStock(context *gin.Context) {
	stocks, err := handler.stockRepo.getAllStocks()
	if err != nil {
		log.Printf("Error getting all stocks: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	txns, err := handler.stockRepo.getAllTxn()
	if err != nil {
		log.Printf("Error getting all stock txn: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	divs, err := handler.stockRepo.getAllDividend()
	if err != nil {
		log.Printf("Error getting all dividends: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	txnByStockName := make(map[string][]Txn)
	for _, t := range txns {
		// append will "append" t to `txnByStockName[t.StockName]` collection, then assign (replace) to txnByStockName[t.StockName]
		txnByStockName[t.StockName] = append(txnByStockName[t.StockName], t)
	}
	divByStockName := make(map[string][]Dividend)
	for _, d := range divs {
		divByStockName[d.StockName] = append(divByStockName[d.StockName], d)
	}
	// length 0 (no item), but capacity = size of stock to avoid resize in iteration
	overviews := make([]Overview, 0, len(stocks))
	for _, s := range stocks {
		stockTxn := txnByStockName[s.StockName]
		stockDiv := divByStockName[s.StockName]
		overview, err := computeOverview(s, stockTxn, stockDiv)
		if err != nil {
			log.Printf("Error compute overview: %v", err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		overviews = append(overviews, overview)
	}
	context.JSON(http.StatusOK, gin.H{"content": overviews})
}

func (handler *handler) getStock(context *gin.Context) {
	stockName := context.Param("stockName")
	stock, err := handler.stockRepo.getStockByStockName(stockName)
	if err != nil {
		log.Printf("Error getting stock: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	txns, err := handler.stockRepo.getStockTxnByStockName(stockName)
	if err != nil {
		log.Printf("Error getting stock: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dividends, err := handler.stockRepo.getDividendByStockName(stock.StockName)
	if err != nil {
		log.Printf("Error getting dividends: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	overview, err := computeOverview(stock, txns, dividends)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, overview)
}

func (handler *handler) getAllStockTransactions(context *gin.Context) {
	stockName := context.Param("stockName")
	stocks, err := handler.stockRepo.getStockTxnByStockName(stockName)
	if err != nil {
		log.Printf("Error getting all stocks summary: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"content": stocks})
}

func (handler *handler) getDividendsByStockName(context *gin.Context) {
	stockName := context.Param("stockName")
	dividends, err := handler.stockRepo.getDividendByStockName(stockName)
	if err != nil {
		log.Printf("Error getting dividends: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"content": dividends})
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
	stockName := context.Param("stockName")
	var req TxnRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn := Txn{
		ID:        uuid.New().String(),
		StockName: stockName,
		TxnDate:   req.TxnDate,
		Unit:      req.Unit,
		UnitPrice: req.UnitPrice,
		BrokerFee: req.BrokerFee,
		TxnType:   req.TxnType,
		Remark:    req.Remark,
	}

	// Calculate total price
	if err := txn.CalculateStockTxnTotalAmount(); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calculation: " + err.Error()})
		return
	}

	if err := handler.stockRepo.createStockTxn(txn); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "created"})
}

func (handler *handler) updateStockTxn(context *gin.Context) {
	stockName := context.Param("stockName")
	txnId := context.Param("txnId")
	var req TxnRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn := Txn{
		ID:        txnId,
		StockName: stockName,
		TxnDate:   req.TxnDate,
		Unit:      req.Unit,
		UnitPrice: req.UnitPrice,
		BrokerFee: req.BrokerFee,
		TxnType:   req.TxnType,
		Remark:    req.Remark,
	}

	// Calculate total price
	if err := txn.CalculateStockTxnTotalAmount(); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calculation: " + err.Error()})
		return
	}

	if err := handler.stockRepo.updateStockTxn(txn); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "updated"})
}

func (handler *handler) createStockDividend(context *gin.Context) {
	var req DividendRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stockName := context.Param("stockName")

	dividend := Dividend{
		StockName:       stockName,
		ExDate:          req.ExDate,
		PaymentDate:     req.PaymentDate,
		StockUnit:       req.StockUnit,
		DividendPerUnit: req.DividendPerUnit,
		TaxPercentage:   req.TaxPercentage,
		Remark:          req.Remark,
	}
	decimalCtx := apd.BaseContext.WithPrecision(14)
	if err := dividend.CalculateDividendTotalAmount(decimalCtx); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calculation: " + err.Error()})
		return
	}
	if err := handler.stockRepo.createDividend(dividend); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "created"})
}

func (handler *handler) updateStockDividend(context *gin.Context) {
	var req DividendRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stockName := context.Param("stockName")

	dividend := Dividend{
		StockName:       stockName,
		ExDate:          req.ExDate,
		PaymentDate:     req.PaymentDate,
		StockUnit:       req.StockUnit,
		DividendPerUnit: req.DividendPerUnit,
		TaxPercentage:   req.TaxPercentage,
		Remark:          req.Remark,
	}
	decimalCtx := apd.BaseContext.WithPrecision(14)
	if err := dividend.CalculateDividendTotalAmount(decimalCtx); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calculation: " + err.Error()})
		return
	}
	if err := handler.stockRepo.updateDividend(dividend); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "updated"})
}

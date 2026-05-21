package fund

import (
	"log"
	"net/http"

	"github.com/cockroachdb/apd/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handler struct {
	repo *repository
}

func newHandler(repo *repository) *handler {
	return &handler{repo: repo}
}

// todo to be refactor to include high level aggregate info
func (h *handler) getAllFunds(c *gin.Context) {
	funds, err := h.repo.getAllFunds()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": funds})
}

// todo to be refactor to include individual aggregate info
func (h *handler) getFund(c *gin.Context) {
	fundCode := c.Param("fundCode")
	fund, err := h.repo.getFundInfo(fundCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, fund)
}

func (h *handler) addFund(c *gin.Context) {
	var req Fund
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	err := h.repo.insertFund(req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *handler) getFundTxnByFundCode(c *gin.Context) {
	fundCode := c.Param("fundCode")
	fundTxns, err := h.repo.getFundTxnByFundCode(fundCode)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": fundTxns})
}

func (h *handler) addFundTxn(c *gin.Context) {
	fundCode := c.Param("fundCode")
	var req TxnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	txn := Txn{
		ID:          uuid.NewString(),
		FundCode:    fundCode,
		TxnDate:     req.TxnDate,
		Unit:        req.Unit,
		UnitPrice:   req.UnitPrice,
		SalesCharge: req.SalesCharge,
		TxnType:     req.TxnType,
		Remark:      req.Remark,
	}

	if err := txn.CalculateTxnTotals(apd.BaseContext.WithPrecision(14)); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if err := h.repo.insertFundTxn(txn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "created"})
}

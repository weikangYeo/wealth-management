package fund

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	repo *repository
}

func newHandler(repo *repository) *handler {
	return &handler{repo: repo}
}

// todo to be refactor to include aggregate info
func (h *handler) getAllFunds(c *gin.Context) {
	funds, err := h.repo.getAllFunds()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": funds})
}

func (h *handler) addFund(c *gin.Context) {
	var req Fund
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	err := h.repo.insertFund(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

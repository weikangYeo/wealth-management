package handlers

import (
	"log"
	"net/http"
	"wealth-management/repository"

	"github.com/gin-gonic/gin"
)

type GoldHandler struct {
	fundRepo repository.GoldRepository
}

// NewGoldHandler it might be fine without but it is a contructor so fundRepo stay private and unmodified when created.
func NewGoldHandler(fundRepo repository.GoldRepository) GoldHandler {
	return GoldHandler{fundRepo: fundRepo}
}

func (handler *GoldHandler) GetAllGoldsTxn(context *gin.Context) {
	funds, err := handler.fundRepo.GetAll()
	if err != nil {
		log.Printf("Error getting funds: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, gin.H{"golds": funds})
}

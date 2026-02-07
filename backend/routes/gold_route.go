package routes

import (
	"database/sql"
	"wealth-management/handlers"
	"wealth-management/repository"

	"github.com/gin-gonic/gin"
)

func registerGoldRoutes(r *gin.Engine, db *sql.DB) {
	goldRepo := repository.NewGoldRepository(db)
	goldHandler := handlers.NewGoldHandler(goldRepo)
	goldRoutes := r.Group("/golds")
	{
		goldRoutes.GET("/", goldHandler.GetAllGoldsTxn)
	}
}

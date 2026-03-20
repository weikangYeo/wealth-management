package gold

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterGoldRoutes(r *gin.Engine, db *sql.DB) {
	goldRepo := newGoldRepository(db)
	goldHandler := newGoldHandler(goldRepo)
	goldRoutes := r.Group("/golds")
	{
		goldRoutes.GET("", goldHandler.getAllGoldsTxn)
		goldRoutes.POST("/bulk-imports", goldHandler.postBulkImportGolds)
	}
}

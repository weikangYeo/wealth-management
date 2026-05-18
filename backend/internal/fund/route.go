package fund

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterFundRoutes(r *gin.Engine, db *sql.DB) {
	repo := newRepository(db)
	h := newHandler(repo)
	route := r.Group("/funds")
	{
		route.GET("", h.getAllFunds)
		route.POST("", h.addFund)
	}
}

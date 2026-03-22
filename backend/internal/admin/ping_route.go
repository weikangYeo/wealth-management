package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterPingRoute(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}

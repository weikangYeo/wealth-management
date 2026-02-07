package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerPingRoute(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}

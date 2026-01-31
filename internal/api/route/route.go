package route

import (
	"net/http"
	"time"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, store *cache.Store) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "UP",
		})
	})

	publicRouter := r.Group("")

	// All Public APIs
	timeout := time.Duration(1) * time.Second

	NewContainerRouter(timeout, publicRouter, store)
	NewGroupRouter(timeout, publicRouter, store)
	NewScheduleRouter(timeout, publicRouter, store)

}

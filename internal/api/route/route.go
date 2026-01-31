package route

import (
	"net/http"
	"time"

	"github.com/bassista/go_spin/internal/app"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, appCtx *app.App) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "UP",
		})
	})

	publicRouter := r.Group("")

	// All Public APIs
	timeout := time.Duration(1) * time.Second

	NewContainerRouter(timeout, publicRouter, appCtx.Cache)
	NewGroupRouter(timeout, publicRouter, appCtx.Cache)
	NewScheduleRouter(timeout, publicRouter, appCtx.Cache)

}

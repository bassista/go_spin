package route

import (
	"net/http"
	"time"

	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/app"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, appCtx *app.App) {
	// Apply CORS middleware globally
	r.Use(middleware.CORSMiddleware(appCtx.Config.Misc.CORSAllowedOrigins))

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
	NewRuntimeRouter(timeout, publicRouter, appCtx.Runtime)

	// UI static files
	NewUIRouter(r)
}

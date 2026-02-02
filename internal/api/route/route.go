package route

import (
	"net/http"

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

	// Serve favicon
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./ui/assets/vite.ico")
	})

	// Redirect root to /ui
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	// All Public APIs
	publicRouter := r.Group("")

	timeout := appCtx.Config.Misc.RequestTimeout

	NewContainerRouter(appCtx.BaseCtx, timeout, publicRouter, appCtx.Cache, appCtx.Runtime)
	NewGroupRouter(timeout, publicRouter, appCtx.Cache)
	NewScheduleRouter(timeout, publicRouter, appCtx.Cache)
	NewRuntimeRouter(appCtx.BaseCtx, timeout, publicRouter, appCtx.Runtime, appCtx.Cache)

	// UI static files
	NewUIRouter(r)
}

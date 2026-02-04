package route

import (
	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/app"
	"github.com/gin-gonic/gin"
)

// NewConfigurationRouter sets up configuration-related routes.
func NewConfigurationRouter(appCtx *app.App, group *gin.RouterGroup) {
	cc := controller.NewConfigurationController(appCtx.Config)
	timeoutMiddleware := middleware.RequestTimeout(appCtx.Config.Server.RequestTimeout)

	group.GET("configuration", timeoutMiddleware, cc.GetConfiguration)
}

package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/config"
	"github.com/gin-gonic/gin"
)

// NewConfigurationRouter sets up configuration-related routes.
func NewConfigurationRouter(timeout time.Duration, group *gin.RouterGroup, cfg *config.Config) {
	group.Use(middleware.RequestTimeout(timeout))

	cc := controller.NewConfigurationController(cfg)

	group.GET("configuration", cc.GetConfiguration)
}

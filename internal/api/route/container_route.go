package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func NewContainerRouter(timeout time.Duration, group *gin.RouterGroup, store cache.ContainerStore) {
	group.Use(middleware.RequestTimeout(timeout))

	cc := controller.NewContainerController(store)

	group.GET("containers", cc.AllContainers)
	group.POST("container", cc.CreateOrUpdateContainer)
	group.DELETE("container/:name", cc.DeleteContainer)
}

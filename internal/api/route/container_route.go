package route

import (
	"context"
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/runtime"
	"github.com/gin-gonic/gin"
)

func NewContainerRouter(ctx context.Context, timeout time.Duration, group *gin.RouterGroup, store cache.ContainerStore, runtime runtime.ContainerRuntime) {
	group.Use(middleware.RequestTimeout(timeout))

	cc := controller.NewContainerController(store, runtime, ctx)

	group.GET("containers", cc.AllContainers)
	group.POST("container", cc.CreateOrUpdateContainer)
	group.DELETE("container/:name", cc.DeleteContainer)
	group.GET("container/:name/ready", cc.Ready)
}

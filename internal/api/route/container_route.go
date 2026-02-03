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
	cc := controller.NewContainerController(ctx, store, runtime)
	timeoutMiddleware := middleware.RequestTimeout(timeout)

	group.GET("containers", timeoutMiddleware, cc.AllContainers)
	group.POST("container", timeoutMiddleware, cc.CreateOrUpdateContainer)
	group.DELETE("container/:name", timeoutMiddleware, cc.DeleteContainer)
	group.GET("container/:name/ready", timeoutMiddleware, cc.Ready)
}

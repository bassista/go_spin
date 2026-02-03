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

func NewGroupRouter(baseCtx context.Context, timeout time.Duration, group *gin.RouterGroup, store cache.GroupStore, rt runtime.ContainerRuntime) {
	gc := controller.NewGroupController(baseCtx, store, rt)
	timeoutMiddleware := middleware.RequestTimeout(timeout)

	group.GET("groups", timeoutMiddleware, gc.AllGroups)
	group.POST("group", timeoutMiddleware, gc.CreateOrUpdateGroup)
	group.DELETE("group/:name", timeoutMiddleware, gc.DeleteGroup)
	group.POST("group/:name/start", timeoutMiddleware, gc.StartGroup)
	group.POST("group/:name/stop", timeoutMiddleware, gc.StopGroup)
}

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
	group.Use(middleware.RequestTimeout(timeout))

	gc := controller.NewGroupController(baseCtx, store, rt)
	group.GET("groups", gc.AllGroups)
	group.POST("group", gc.CreateOrUpdateGroup)
	group.DELETE("group/:name", gc.DeleteGroup)
	group.POST("group/:name/start", gc.StartGroup)
	group.POST("group/:name/stop", gc.StopGroup)
}

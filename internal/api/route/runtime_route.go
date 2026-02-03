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

func NewRuntimeRouter(baseCtx context.Context, timeout time.Duration, group *gin.RouterGroup, rt runtime.ContainerRuntime, store cache.ContainerStore) {
	rc := controller.NewRuntimeController(baseCtx, rt, store)

	// Apply default timeout middleware to most routes
	defaultTimeout := middleware.RequestTimeout(timeout)

	group.GET("runtime/:name/status", defaultTimeout, rc.IsRunning)
	group.POST("runtime/:name/start", defaultTimeout, rc.StartContainer)
	group.POST("runtime/:name/stop", defaultTimeout, rc.StopContainer)
	group.GET("runtime/containers", defaultTimeout, rc.ListContainers)
	group.GET("start/:name", defaultTimeout, rc.WaitingPage)

	// Stats endpoint needs a longer timeout (30s) since it queries all containers
	group.GET("runtime/stats", middleware.RequestTimeout(30*time.Second), rc.AllStats)
}

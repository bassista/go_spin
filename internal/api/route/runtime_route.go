package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/runtime"
	"github.com/gin-gonic/gin"
)

func NewRuntimeRouter(timeout time.Duration, group *gin.RouterGroup, rt runtime.ContainerRuntime, store cache.ContainerStore) {
	group.Use(middleware.RequestTimeout(timeout))

	rc := controller.NewRuntimeController(rt, store)

	group.GET("runtime/:name/status", rc.IsRunning)
	group.POST("runtime/:name/start", rc.StartContainer)
	group.POST("runtime/:name/stop", rc.StopContainer)
}

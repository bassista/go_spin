package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/runtime"
	"github.com/gin-gonic/gin"
)

func NewRuntimeRouter(timeout time.Duration, group *gin.RouterGroup, rt runtime.ContainerRuntime) {
	group.Use(middleware.RequestTimeout(timeout))

	rc := controller.NewRuntimeController(rt)

	group.GET("runtime/:name/status", rc.IsRunning)
	group.POST("runtime/:name/start", rc.StartContainer)
	group.POST("runtime/:name/stop", rc.StopContainer)
}

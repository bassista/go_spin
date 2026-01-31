package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func NewGroupRouter(timeout time.Duration, group *gin.RouterGroup, store cache.GroupStore) {
	group.Use(middleware.RequestTimeout(timeout))

	gc := controller.NewGroupController(store)
	group.GET("groups", gc.AllGroups)
	group.POST("group", gc.CreateOrUpdateGroup)
	group.DELETE("group/:name", gc.DeleteGroup)
}

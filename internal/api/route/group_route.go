package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func NewGroupRouter(timeout time.Duration, group *gin.RouterGroup, store *cache.Store) {
	gc := controller.NewGroupController(store)
	group.GET("groups", gc.AllGroups)
}

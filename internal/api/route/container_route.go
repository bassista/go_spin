package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func NewContainerRouter(timeout time.Duration, group *gin.RouterGroup, store *cache.Store) {

	cc := controller.NewContainerController(store)

	group.GET("containers", cc.AllContainers)
	group.POST("container", cc.CreateOrUpdateContainer)
}

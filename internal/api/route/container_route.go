package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func NewContainerRouter(timeout time.Duration, group *gin.RouterGroup, store *cache.Store, validator *validator.Validate) {

	cc := controller.NewContainerController(store, validator)

	group.GET("containers", cc.AllContainers)
	group.POST("container", cc.CreateOrUpdateContainer)
}

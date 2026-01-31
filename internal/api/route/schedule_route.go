package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func NewScheduleRouter(timeout time.Duration, group *gin.RouterGroup, store *cache.Store) {
	sc := controller.NewScheduleController(store)
	group.GET("schedules", sc.AllSchedules)
}

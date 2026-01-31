package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func NewScheduleRouter(timeout time.Duration, group *gin.RouterGroup, store cache.ScheduleStore) {
	sc := controller.NewScheduleController(store)
	group.GET("schedules", sc.AllSchedules)
	group.POST("schedule", sc.CreateOrUpdateSchedule)
	group.DELETE("schedule/:id", sc.DeleteSchedule)
}

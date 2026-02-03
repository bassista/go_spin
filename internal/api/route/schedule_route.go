package route

import (
	"time"

	"github.com/bassista/go_spin/internal/api/controller"
	"github.com/bassista/go_spin/internal/api/middleware"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

func NewScheduleRouter(timeout time.Duration, group *gin.RouterGroup, store cache.ScheduleStore) {
	sc := controller.NewScheduleController(store)
	timeoutMiddleware := middleware.RequestTimeout(timeout)

	group.GET("schedules", timeoutMiddleware, sc.AllSchedules)
	group.POST("schedule", timeoutMiddleware, sc.CreateOrUpdateSchedule)
	group.DELETE("schedule/:id", timeoutMiddleware, sc.DeleteSchedule)
}

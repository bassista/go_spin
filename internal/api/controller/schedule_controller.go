package controller

import (
	"net/http"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

// ScheduleController exposes schedule-related handlers.
type ScheduleController struct {
	store *cache.Store
}

// NewScheduleController builds a ScheduleController with cache store.
func NewScheduleController(store *cache.Store) *ScheduleController {
	return &ScheduleController{store: store}
}

// AllSchedules returns all schedules from cache.
func (sc *ScheduleController) AllSchedules(c *gin.Context) {
	data, err := sc.store.Snapshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read cache"})
		return
	}
	c.JSON(http.StatusOK, data.Schedules)
}

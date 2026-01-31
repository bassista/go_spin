package controller

import (
	"errors"
	"net/http"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ScheduleController exposes schedule-related handlers.
type ScheduleController struct {
	store     cache.ScheduleStore
	validator *validator.Validate
}

// NewScheduleController builds a ScheduleController with cache store.
func NewScheduleController(store cache.ScheduleStore) *ScheduleController {
	v := validator.New()
	return &ScheduleController{store: store, validator: v}
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

// CreateOrUpdateSchedule upserts a schedule and returns the full list.
// Persistence is handled by the scheduled persistence goroutine.
func (sc *ScheduleController) CreateOrUpdateSchedule(c *gin.Context) {
	var schedule repository.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if sc.validator != nil {
		if err := sc.validator.Struct(schedule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	updatedDoc, err := sc.store.AddSchedule(schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	c.JSON(http.StatusOK, updatedDoc.Schedules)
}

func (sc *ScheduleController) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing schedule id"})
		return
	}

	updatedDoc, err := sc.store.RemoveSchedule(id)
	if err != nil {
		if errors.Is(err, cache.ErrScheduleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	c.JSON(http.StatusOK, updatedDoc.Schedules)
}

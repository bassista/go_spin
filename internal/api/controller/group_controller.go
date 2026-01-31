package controller

import (
	"errors"
	"net/http"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GroupController exposes group-related handlers.
type GroupController struct {
	store     cache.GroupStore
	validator *validator.Validate
}

// NewGroupController builds a GroupController with cache store.
func NewGroupController(store cache.GroupStore) *GroupController {
	v := validator.New()
	return &GroupController{store: store, validator: v}
}

// AllGroups returns all groups from cache.
func (gc *GroupController) AllGroups(c *gin.Context) {
	data, err := gc.store.Snapshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read cache"})
		return
	}
	c.JSON(http.StatusOK, data.Groups)
}

// CreateOrUpdateGroup upserts a group and returns the full list.
// Persistence is handled by the scheduled persistence goroutine.
func (gc *GroupController) CreateOrUpdateGroup(c *gin.Context) {
	var group repository.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if gc.validator != nil {
		if err := gc.validator.Struct(group); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	updatedDoc, err := gc.store.AddGroup(group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	c.JSON(http.StatusOK, updatedDoc.Groups)
}

func (gc *GroupController) DeleteGroup(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing group name"})
		return
	}

	updatedDoc, err := gc.store.RemoveGroup(name)
	if err != nil {
		if errors.Is(err, cache.ErrGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	c.JSON(http.StatusOK, updatedDoc.Groups)
}

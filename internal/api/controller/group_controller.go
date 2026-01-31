package controller

import (
	"net/http"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
)

// GroupController exposes group-related handlers.
type GroupController struct {
	store *cache.Store
}

// NewGroupController builds a GroupController with cache store.
func NewGroupController(store *cache.Store) *GroupController {
	return &GroupController{store: store}
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

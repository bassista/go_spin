package controller

import (
	"errors"
	"net/http"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/logger"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GroupController handles group-related HTTP endpoints using the generic CRUD controller.
type GroupController struct {
	crud *CrudController[repository.Group]
}

// NewGroupController creates a new GroupController with the given cache store.
func NewGroupController(store cache.GroupStore) *GroupController {
	v := validator.New()
	service := &GroupCrudService{Store: store}
	validator := &GroupCrudValidator{validator: v}

	return &GroupController{
		crud: &CrudController[repository.Group]{
			Service:   service,
			Validator: validator,
		},
	}
}

// AllGroups handles GET /groups - returns all groups.
func (gc *GroupController) AllGroups(c *gin.Context) {
	logger.WithComponent("group-controller").Debugf("GET /groups handler called")
	gc.crud.GetAll(c)
}

// CreateOrUpdateGroup handles POST /group - creates or updates a group.
func (gc *GroupController) CreateOrUpdateGroup(c *gin.Context) {
	logger.WithComponent("group-controller").Debugf("POST /group handler called")
	gc.crud.CreateOrUpdate(c)
}

// DeleteGroup handles DELETE /group/:name - deletes a group by name.
func (gc *GroupController) DeleteGroup(c *gin.Context) {
	name := c.Param("name")
	logger.WithComponent("group-controller").Debugf("DELETE /group/%s handler called", name)
	if name == "" {
		logger.WithComponent("group-controller").Debugf("delete group: missing name parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing group name"})
		return
	}

	items, err := gc.crud.Service.Remove(name)
	if err != nil {
		if errors.Is(err, cache.ErrGroupNotFound) {
			logger.WithComponent("group-controller").Debugf("delete group %s: not found", name)
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		logger.WithComponent("group-controller").Errorf("delete group %s: cache error: %v", name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	logger.WithComponent("group-controller").Debugf("group %s deleted successfully", name)
	c.JSON(http.StatusOK, items)
}

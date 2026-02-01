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

// ContainerController handles container-related HTTP endpoints using the generic CRUD controller.
type ContainerController struct {
	crud *CrudController[repository.Container]
}

// NewContainerController creates a new ContainerController with the given cache store.
func NewContainerController(store cache.ContainerStore) *ContainerController {
	v := validator.New()
	service := &ContainerCrudService{Store: store}
	validator := &ContainerCrudValidator{validator: v}

	return &ContainerController{
		crud: &CrudController[repository.Container]{
			Service:   service,
			Validator: validator,
		},
	}
}

// AllContainers handles GET /containers - returns all containers.
func (cc *ContainerController) AllContainers(c *gin.Context) {
	logger.WithComponent("container-controller").Debugf("GET /containers handler called")
	cc.crud.GetAll(c)
}

// CreateOrUpdateContainer handles POST /container - creates or updates a container.
func (cc *ContainerController) CreateOrUpdateContainer(c *gin.Context) {
	logger.WithComponent("container-controller").Debugf("POST /container handler called")
	cc.crud.CreateOrUpdate(c)
}

// DeleteContainer handles DELETE /container/:name - deletes a container by name.
func (cc *ContainerController) DeleteContainer(c *gin.Context) {
	name := c.Param("name")
	logger.WithComponent("container-controller").Debugf("DELETE /container/%s handler called", name)
	if name == "" {
		logger.WithComponent("container-controller").Debugf("delete container: missing name parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing container name"})
		return
	}

	items, err := cc.crud.Service.Remove(name)
	if err != nil {
		if errors.Is(err, cache.ErrContainerNotFound) {
			logger.WithComponent("container-controller").Debugf("delete container %s: not found", name)
			c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
			return
		}
		logger.WithComponent("container-controller").Errorf("delete container %s: cache error: %v", name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	logger.WithComponent("container-controller").Debugf("container %s deleted successfully", name)
	c.JSON(http.StatusOK, items)
}

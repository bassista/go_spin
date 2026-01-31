package controller

import (
	"errors"
	"net/http"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ContainerController struct {
	store     cache.ContainerStore
	validator *validator.Validate
}

func NewContainerController(store cache.ContainerStore) *ContainerController {
	validator := validator.New()
	return &ContainerController{store: store, validator: validator}
}

func (cc *ContainerController) AllContainers(c *gin.Context) {
	data, err := cc.store.Snapshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read cache"})
		return
	}

	c.JSON(http.StatusOK, data.Containers)
}

// CreateOrUpdateContainer upserts a container and returns the full list.
// Persistence is handled by the scheduled persistence goroutine.
func (cc *ContainerController) CreateOrUpdateContainer(c *gin.Context) {
	var container repository.Container
	if err := c.ShouldBindJSON(&container); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if cc.validator != nil {
		if err := cc.validator.Struct(container); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	updatedDoc, err := cc.store.AddContainer(container)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	c.JSON(http.StatusOK, updatedDoc.Containers)
}

func (cc *ContainerController) DeleteContainer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing container name"})
		return
	}

	updatedDoc, err := cc.store.RemoveContainer(name)
	if err != nil {
		if errors.Is(err, cache.ErrContainerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache"})
		return
	}

	c.JSON(http.StatusOK, updatedDoc.Containers)
}

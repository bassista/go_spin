package controller

import (
	"net/http"
	"strings"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/runtime"
	"github.com/gin-gonic/gin"
)

type RuntimeController struct {
	runtime        runtime.ContainerRuntime
	containerStore cache.ContainerStore
}

func NewRuntimeController(rt runtime.ContainerRuntime, store cache.ContainerStore) *RuntimeController {
	return &RuntimeController{
		runtime:        rt,
		containerStore: store,
	}
}

// IsRunning checks if a container is currently running.
func (rc *RuntimeController) IsRunning(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing container name"})
		return
	}

	// Check if container exists in cache
	doc, err := rc.containerStore.Snapshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read container list"})
		return
	}

	containerExists := false
	for _, container := range doc.Containers {
		if container.Name == name {
			containerExists = true
			break
		}
	}
	if !containerExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
		return
	}

	running, err := rc.runtime.IsRunning(c.Request.Context(), name)
	if err != nil {
		// Check if error is "container not found"
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    name,
		"running": running,
	})
}

// StartContainer starts a container by name.
func (rc *RuntimeController) StartContainer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing container name"})
		return
	}

	// Check if container exists in cache
	doc, err := rc.containerStore.Snapshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read container list"})
		return
	}

	containerExists := false
	for _, container := range doc.Containers {
		if container.Name == name {
			containerExists = true
			break
		}
	}
	if !containerExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
		return
	}

	if err := rc.runtime.Start(c.Request.Context(), name); err != nil {
		// Check if error is "container not found"
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    name,
		"message": "container started",
	})
}

// StopContainer stops a container by name.
func (rc *RuntimeController) StopContainer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing container name"})
		return
	}

	// Check if container exists in cache
	doc, err := rc.containerStore.Snapshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read container list"})
		return
	}

	containerExists := false
	for _, container := range doc.Containers {
		if container.Name == name {
			containerExists = true
			break
		}
	}
	if !containerExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
		return
	}

	if err := rc.runtime.Stop(c.Request.Context(), name); err != nil {
		// Check if error is "container not found"
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    name,
		"message": "container stopped",
	})
}

package controller

import (
	"net/http"

	"github.com/bassista/go_spin/internal/runtime"
	"github.com/gin-gonic/gin"
)

type RuntimeController struct {
	runtime runtime.ContainerRuntime
}

func NewRuntimeController(rt runtime.ContainerRuntime) *RuntimeController {
	return &RuntimeController{runtime: rt}
}

// IsRunning checks if a container is currently running.
func (rc *RuntimeController) IsRunning(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing container name"})
		return
	}

	running, err := rc.runtime.IsRunning(c.Request.Context(), name)
	if err != nil {
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

	if err := rc.runtime.Start(c.Request.Context(), name); err != nil {
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

	if err := rc.runtime.Stop(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    name,
		"message": "container stopped",
	})
}

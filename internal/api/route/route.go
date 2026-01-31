package route

import (
	"fmt"
	"time"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func SetupRoutes(timeout time.Duration, r *gin.Engine, store *cache.Store, validator *validator.Validate) {
	fmt.Println("DEBUG: SetupRoutes called!")

	publicRouter := r.Group("")
	// All Public APIs
	NewContainerRouter(timeout, publicRouter, store, validator)
	NewGroupRouter(timeout, publicRouter, store)
	NewScheduleRouter(timeout, publicRouter, store)

}

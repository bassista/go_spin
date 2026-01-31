package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewUIRouter sets up routes to serve the UI static files under /ui.
// It serves index.html for the root and any sub-paths (SPA routing).
func NewUIRouter(r *gin.Engine) {
	// Serve static assets (JS, CSS, images)
	r.Static("/ui/assets", "./ui/assets")

	// Serve index.html for the /ui root
	r.GET("/ui", func(c *gin.Context) {
		c.File("./ui/index.html")
	})

	// Serve index.html for any sub-path under /ui (SPA client-side routing)
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path

		// Only handle /ui/* paths, return 404 for others
		if p == "/ui" || strings.HasPrefix(p, "/ui/") {
			c.File("./ui/index.html")
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}

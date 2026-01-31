package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a Gin middleware that handles CORS preflight and headers.
// allowedOrigins is a comma-separated list of allowed origins, or "*" for all.
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	// Pre-parse allowed origins for efficiency
	var allowAll bool
	var originSet map[string]struct{}

	if allowedOrigins == "*" {
		allowAll = true
	} else {
		originSet = make(map[string]struct{})
		for _, o := range strings.Split(allowedOrigins, ",") {
			o = strings.TrimSpace(o)
			if o == "" {
				continue
			}
			originSet[o] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Determine which origin to return
		var allowOrigin string
		if allowAll {
			allowOrigin = "*"
		} else if origin != "" {
			if _, ok := originSet[origin]; ok {
				allowOrigin = origin
			}
		}

		// If origin not allowed, skip CORS headers (browser will block)
		if allowOrigin == "" {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Max-Age", "86400")

		// For preflight, echo requested headers if present; otherwise use defaults
		reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
		if strings.TrimSpace(reqHeaders) != "" {
			c.Header("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		}

		// Only set credentials if not using wildcard origin
		if allowOrigin != "*" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

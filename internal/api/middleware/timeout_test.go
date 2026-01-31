package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRequestTimeout_ZeroDuration(t *testing.T) {
	r := gin.New()
	r.Use(RequestTimeout(0))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequestTimeout_NegativeDuration(t *testing.T) {
	r := gin.New()
	r.Use(RequestTimeout(-1 * time.Second))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequestTimeout_RequestCompletesBeforeTimeout(t *testing.T) {
	r := gin.New()
	r.Use(RequestTimeout(5 * time.Second))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequestTimeout_ContextHasDeadline(t *testing.T) {
	r := gin.New()
	r.Use(RequestTimeout(5 * time.Second))

	var hasDeadline bool
	r.GET("/test", func(c *gin.Context) {
		_, hasDeadline = c.Request.Context().Deadline()
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !hasDeadline {
		t.Error("expected context to have deadline")
	}
}

func TestRequestTimeout_TimeoutTriggered(t *testing.T) {
	r := gin.New()
	r.Use(RequestTimeout(50 * time.Millisecond))
	r.GET("/test", func(c *gin.Context) {
		// Simulate slow operation
		select {
		case <-time.After(200 * time.Millisecond):
			c.String(http.StatusOK, "ok")
		case <-c.Request.Context().Done():
			// Handler respects context cancellation
			return
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// The middleware checks if context timed out AND nothing was written
	// Since handler returned without writing, should get 504
	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status 504 Gateway Timeout, got %d", w.Code)
	}
}

func TestRequestTimeout_HandlerWritesBeforeTimeout(t *testing.T) {
	r := gin.New()
	r.Use(RequestTimeout(100 * time.Millisecond))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
		// Sleep after writing - timeout shouldn't override response
		time.Sleep(150 * time.Millisecond)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Response was already written, should be 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 (already written), got %d", w.Code)
	}
}

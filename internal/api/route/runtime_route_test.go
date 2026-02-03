package route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/bassista/go_spin/internal/repository"
	"github.com/bassista/go_spin/internal/runtime"
	"github.com/gin-gonic/gin"
)

// mockContainerRuntime implements runtime.ContainerRuntime for testing
type mockContainerRuntime struct {
	mu           sync.RWMutex
	statsDelay   time.Duration
	statsCtxUsed context.Context
}

func (m *mockContainerRuntime) IsRunning(ctx context.Context, name string) (bool, error) {
	return true, nil
}

func (m *mockContainerRuntime) Start(ctx context.Context, name string) error {
	return nil
}

func (m *mockContainerRuntime) Stop(ctx context.Context, name string) error {
	return nil
}

func (m *mockContainerRuntime) ListContainers(ctx context.Context) ([]string, error) {
	return []string{"test-container"}, nil
}

func (m *mockContainerRuntime) Stats(ctx context.Context, containerName string) (runtime.ContainerStats, error) {
	m.mu.Lock()
	m.statsCtxUsed = ctx
	delay := m.statsDelay
	m.mu.Unlock()

	if delay > 0 {
		select {
		case <-time.After(delay):
			return runtime.ContainerStats{CPUPercent: 10.0, MemoryMB: 100.0}, nil
		case <-ctx.Done():
			return runtime.ContainerStats{}, ctx.Err()
		}
	}
	return runtime.ContainerStats{CPUPercent: 10.0, MemoryMB: 100.0}, nil
}

// mockContainerStore implements cache.ContainerStore for testing
type mockContainerStore struct{}

func (m *mockContainerStore) Snapshot() (repository.DataDocument, error) {
	return repository.DataDocument{
		Containers: []repository.Container{
			{Name: "test-container"},
		},
	}, nil
}

func (m *mockContainerStore) AddContainer(container repository.Container) (repository.DataDocument, error) {
	return repository.DataDocument{}, nil
}

func (m *mockContainerStore) RemoveContainer(name string) (repository.DataDocument, error) {
	return repository.DataDocument{}, nil
}

func TestRuntimeRoute_StatsEndpointHasLongerTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRT := &mockContainerRuntime{
		statsDelay: 500 * time.Millisecond, // Stats takes 500ms
	}
	mockStore := &mockContainerStore{}

	r := gin.New()
	group := r.Group("/api")

	// Default timeout is very short (100ms), but stats should have 30s
	NewRuntimeRouter(context.Background(), 100*time.Millisecond, 30*time.Second, group, mockRT, mockStore)

	// Test that stats endpoint succeeds even though it takes longer than default timeout
	req, _ := http.NewRequest(http.MethodGet, "/api/runtime/stats", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRuntimeRoute_DefaultTimeoutAppliedToOtherRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRT := &mockContainerRuntime{}
	mockStore := &mockContainerStore{}

	r := gin.New()
	group := r.Group("/api")

	// Use a short timeout
	NewRuntimeRouter(context.Background(), 50*time.Millisecond, 30*time.Second, group, mockRT, mockStore)

	// Test that containers endpoint gets the default timeout context
	req, _ := http.NewRequest(http.MethodGet, "/api/runtime/containers", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should succeed since ListContainers is fast
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRuntimeRoute_StatsTimeoutIsIndependentFromDefaultTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// This test verifies that the stats endpoint timeout (30s) is NOT limited
	// by the default timeout. We use a very short default timeout and a stats
	// operation that takes longer than that but less than 30s.

	mockRT := &mockContainerRuntime{
		statsDelay: 200 * time.Millisecond, // Stats takes 200ms
	}
	mockStore := &mockContainerStore{}

	r := gin.New()
	group := r.Group("/api")

	// Default timeout is 50ms, much shorter than stats delay
	// If stats used the default timeout, it would fail
	NewRuntimeRouter(context.Background(), 50*time.Millisecond, 30*time.Second, group, mockRT, mockStore)

	req, _ := http.NewRequest(http.MethodGet, "/api/runtime/stats", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Stats should succeed because it has its own 30s timeout, not 50ms
	if w.Code != http.StatusOK {
		t.Errorf("stats endpoint should have succeeded with its own timeout, got status %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRuntimeRoute_StatsContextDeadlineIsApproximately30Seconds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedDeadline time.Time
	var hasDeadline bool

	mockRT := &mockContainerRuntime{}
	mockStore := &mockContainerStore{}

	// Create a custom handler to capture the context deadline
	r := gin.New()
	group := r.Group("/api")

	NewRuntimeRouter(context.Background(), 100*time.Millisecond, 30*time.Second, group, mockRT, mockStore)

	// We need to intercept the context. Let's modify our approach:
	// Instead, we verify the behavior by checking that a 200ms operation succeeds
	// when default timeout is 50ms (proving stats has its own timeout)

	// This is already covered by TestRuntimeRoute_StatsTimeoutIsIndependentFromDefaultTimeout
	// Let's verify the deadline is set correctly by checking the context
	mockRT2 := &mockContainerRuntime{}

	r2 := gin.New()
	var capturedCtx context.Context
	r2.GET("/test-stats", func(c *gin.Context) {
		capturedCtx = c.Request.Context()
		capturedDeadline, hasDeadline = capturedCtx.Deadline()
		c.JSON(200, gin.H{"ok": true})
	})

	// Apply our middleware directly
	r3 := gin.New()
	group3 := r3.Group("/api")
	NewRuntimeRouter(context.Background(), 100*time.Millisecond, 30*time.Second, group3, mockRT2, mockStore)

	req, _ := http.NewRequest(http.MethodGet, "/api/runtime/stats", nil)
	w := httptest.NewRecorder()

	startTime := time.Now()
	r3.ServeHTTP(w, req)

	// Get the context from mock runtime
	mockRT2.mu.Lock()
	if mockRT2.statsCtxUsed != nil {
		capturedDeadline, hasDeadline = mockRT2.statsCtxUsed.Deadline()
	}
	mockRT2.mu.Unlock()

	if !hasDeadline {
		t.Error("expected stats context to have a deadline")
		return
	}

	// Deadline should be approximately 30 seconds from now (with some tolerance)
	expectedDeadline := startTime.Add(30 * time.Second)
	tolerance := 2 * time.Second

	if capturedDeadline.Before(expectedDeadline.Add(-tolerance)) || capturedDeadline.After(expectedDeadline.Add(tolerance)) {
		t.Errorf("stats context deadline should be ~30s, got deadline at %v (expected around %v)", capturedDeadline, expectedDeadline)
	}
}

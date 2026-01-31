package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockContainerRuntime implements runtime.ContainerRuntime for testing
type mockContainerRuntime struct {
	runningContainers map[string]bool
	startErr          error
	stopErr           error
	isRunningErr      error
}

func newMockRuntime() *mockContainerRuntime {
	return &mockContainerRuntime{
		runningContainers: make(map[string]bool),
	}
}

func (m *mockContainerRuntime) IsRunning(ctx context.Context, name string) (bool, error) {
	if m.isRunningErr != nil {
		return false, m.isRunningErr
	}
	return m.runningContainers[name], nil
}

func (m *mockContainerRuntime) Start(ctx context.Context, name string) error {
	if m.startErr != nil {
		return m.startErr
	}
	m.runningContainers[name] = true
	return nil
}

func (m *mockContainerRuntime) Stop(ctx context.Context, name string) error {
	if m.stopErr != nil {
		return m.stopErr
	}
	m.runningContainers[name] = false
	return nil
}

func TestRuntimeController_IsRunning_Success(t *testing.T) {
	rt := newMockRuntime()
	rt.runningContainers["my-container"] = true

	rc := NewRuntimeController(rt)

	r := gin.New()
	r.GET("/runtime/:name/status", rc.IsRunning)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/status", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["name"] != "my-container" {
		t.Errorf("expected name 'my-container', got %v", resp["name"])
	}
	if resp["running"] != true {
		t.Errorf("expected running true, got %v", resp["running"])
	}
}

func TestRuntimeController_IsRunning_NotRunning(t *testing.T) {
	rt := newMockRuntime()
	rt.runningContainers["stopped-container"] = false

	rc := NewRuntimeController(rt)

	r := gin.New()
	r.GET("/runtime/:name/status", rc.IsRunning)

	req := httptest.NewRequest(http.MethodGet, "/runtime/stopped-container/status", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["running"] != false {
		t.Errorf("expected running false, got %v", resp["running"])
	}
}

func TestRuntimeController_IsRunning_MissingName(t *testing.T) {
	rt := newMockRuntime()
	rc := NewRuntimeController(rt)

	r := gin.New()
	// Test with empty name param - controller validates and returns 400
	r.GET("/runtime/:name/status", rc.IsRunning)

	// Request with empty path segment - Gin still matches with empty :name
	req := httptest.NewRequest(http.MethodGet, "/runtime//status", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Controller returns 400 for empty name
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRuntimeController_IsRunning_RuntimeError(t *testing.T) {
	rt := newMockRuntime()
	rt.isRunningErr = errors.New("docker connection failed")

	rc := NewRuntimeController(rt)

	r := gin.New()
	r.GET("/runtime/:name/status", rc.IsRunning)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/status", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_StartContainer_Success(t *testing.T) {
	rt := newMockRuntime()
	rc := NewRuntimeController(rt)

	r := gin.New()
	r.POST("/runtime/:name/start", rc.StartContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/my-container/start", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["name"] != "my-container" {
		t.Errorf("expected name 'my-container', got %v", resp["name"])
	}
	if resp["message"] != "container started" {
		t.Errorf("expected message 'container started', got %v", resp["message"])
	}

	// Verify container is now running in mock
	if !rt.runningContainers["my-container"] {
		t.Error("expected container to be marked as running in mock")
	}
}

func TestRuntimeController_StartContainer_MissingName(t *testing.T) {
	rt := newMockRuntime()
	rc := NewRuntimeController(rt)

	r := gin.New()
	// Test with empty name param - controller validates and returns 400
	r.POST("/runtime/:name/start", rc.StartContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime//start", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Controller returns 400 for empty name
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRuntimeController_StartContainer_RuntimeError(t *testing.T) {
	rt := newMockRuntime()
	rt.startErr = errors.New("container not found")

	rc := NewRuntimeController(rt)

	r := gin.New()
	r.POST("/runtime/:name/start", rc.StartContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/nonexistent/start", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_StopContainer_Success(t *testing.T) {
	rt := newMockRuntime()
	rt.runningContainers["my-container"] = true

	rc := NewRuntimeController(rt)

	r := gin.New()
	r.POST("/runtime/:name/stop", rc.StopContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/my-container/stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["name"] != "my-container" {
		t.Errorf("expected name 'my-container', got %v", resp["name"])
	}
	if resp["message"] != "container stopped" {
		t.Errorf("expected message 'container stopped', got %v", resp["message"])
	}

	// Verify container is now stopped in mock
	if rt.runningContainers["my-container"] {
		t.Error("expected container to be marked as stopped in mock")
	}
}

func TestRuntimeController_StopContainer_MissingName(t *testing.T) {
	rt := newMockRuntime()
	rc := NewRuntimeController(rt)

	r := gin.New()
	// Test with empty name param - controller validates and returns 400
	r.POST("/runtime/:name/stop", rc.StopContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime//stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Controller returns 400 for empty name
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRuntimeController_StopContainer_RuntimeError(t *testing.T) {
	rt := newMockRuntime()
	rt.stopErr = errors.New("container already stopped")

	rc := NewRuntimeController(rt)

	r := gin.New()
	r.POST("/runtime/:name/stop", rc.StopContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/my-container/stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_FullLifecycle(t *testing.T) {
	rt := newMockRuntime()
	rc := NewRuntimeController(rt)

	r := gin.New()
	r.GET("/runtime/:name/status", rc.IsRunning)
	r.POST("/runtime/:name/start", rc.StartContainer)
	r.POST("/runtime/:name/stop", rc.StopContainer)

	containerName := "lifecycle-test"

	// 1. Check initial status (should be not running)
	req := httptest.NewRequest(http.MethodGet, "/runtime/"+containerName+"/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["running"] != false {
		t.Errorf("expected container initially not running")
	}

	// 2. Start container
	req = httptest.NewRequest(http.MethodPost, "/runtime/"+containerName+"/start", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("failed to start container: %d", w.Code)
	}

	// 3. Check status (should be running)
	req = httptest.NewRequest(http.MethodGet, "/runtime/"+containerName+"/status", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["running"] != true {
		t.Errorf("expected container to be running after start")
	}

	// 4. Stop container
	req = httptest.NewRequest(http.MethodPost, "/runtime/"+containerName+"/stop", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("failed to stop container: %d", w.Code)
	}

	// 5. Check status (should be stopped)
	req = httptest.NewRequest(http.MethodGet, "/runtime/"+containerName+"/status", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["running"] != false {
		t.Errorf("expected container to be stopped after stop")
	}
}

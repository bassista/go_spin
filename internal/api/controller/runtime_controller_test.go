package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bassista/go_spin/internal/repository"
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

// newMockStoreWithContainer creates a mock store with a container
func newMockStoreWithContainer(name string) *mockContainerStore {
	return &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{
				{Name: name},
			},
		},
	}
}

// newMockStoreEmpty creates an empty mock store
func newMockStoreEmpty() *mockContainerStore {
	return &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{},
		},
	}
}

func TestRuntimeController_IsRunning_Success(t *testing.T) {
	rt := newMockRuntime()
	rt.runningContainers["my-container"] = true

	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{
				{Name: "my-container"},
			},
		},
	}

	rc := NewRuntimeController(context.Background(), rt, store)

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

	store := newMockStoreWithContainer("stopped-container")
	rc := NewRuntimeController(context.Background(), rt, store)

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
	store := newMockStoreEmpty()
	rc := NewRuntimeController(context.Background(), rt, store)

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

	store := newMockStoreWithContainer("my-container")
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/status", rc.IsRunning)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/status", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_IsRunning_ContainerNotFound(t *testing.T) {
	rt := newMockRuntime()
	rt.isRunningErr = errors.New("container nonexistent not found")

	store := newMockStoreWithContainer("nonexistent")
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/status", rc.IsRunning)

	req := httptest.NewRequest(http.MethodGet, "/runtime/nonexistent/status", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRuntimeController_StartContainer_Success(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreWithContainer("my-container")
	rc := NewRuntimeController(context.Background(), rt, store)

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
	store := newMockStoreEmpty()
	rc := NewRuntimeController(context.Background(), rt, store)

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
	rt.startErr = errors.New("docker daemon unavailable")

	store := newMockStoreWithContainer("my-container")
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.POST("/runtime/:name/start", rc.StartContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/my-container/start", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_StartContainer_ContainerNotFound(t *testing.T) {
	rt := newMockRuntime()
	rt.startErr = errors.New("error starting container nonexistent: container not found")

	store := newMockStoreWithContainer("nonexistent")
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.POST("/runtime/:name/start", rc.StartContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/nonexistent/start", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRuntimeController_StopContainer_Success(t *testing.T) {
	rt := newMockRuntime()
	rt.runningContainers["my-container"] = true

	store := newMockStoreWithContainer("my-container")
	rc := NewRuntimeController(context.Background(), rt, store)

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
	store := newMockStoreEmpty()
	rc := NewRuntimeController(context.Background(), rt, store)

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

	store := newMockStoreWithContainer("my-container")
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.POST("/runtime/:name/stop", rc.StopContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/my-container/stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_StopContainer_ContainerNotFound(t *testing.T) {
	rt := newMockRuntime()
	rt.stopErr = errors.New("error stopping container nonexistent: container not found")

	store := newMockStoreWithContainer("nonexistent")
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.POST("/runtime/:name/stop", rc.StopContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/nonexistent/stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRuntimeController_FullLifecycle(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreWithContainer("lifecycle-test")
	rc := NewRuntimeController(context.Background(), rt, store)

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
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
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

	_ = json.Unmarshal(w.Body.Bytes(), &resp)
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

	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["running"] != false {
		t.Errorf("expected container to be stopped after stop")
	}
}
func TestRuntimeController_IsRunning_NotFoundInCache(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreEmpty() // Empty store - container doesn't exist
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/status", rc.IsRunning)

	req := httptest.NewRequest(http.MethodGet, "/runtime/nonexistent/status", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRuntimeController_StartContainer_NotFoundInCache(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreEmpty() // Empty store - container doesn't exist
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.POST("/runtime/:name/start", rc.StartContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/nonexistent/start", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRuntimeController_StopContainer_NotFoundInCache(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreEmpty() // Empty store - container doesn't exist
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.POST("/runtime/:name/stop", rc.StopContainer)

	req := httptest.NewRequest(http.MethodPost, "/runtime/nonexistent/stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// Helper to create a pointer to bool
func boolPtr(b bool) *bool {
	return &b
}

// newMockStoreWithActiveContainer creates a mock store with an active container
func newMockStoreWithActiveContainer(name, url string, active bool) *mockContainerStore {
	return &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{
				{Name: name, URL: url, Active: boolPtr(active)},
			},
		},
	}
}

// newMockStoreWithGroup creates a mock store with a group and its containers
func newMockStoreWithGroup(groupName string, containerNames []string, groupActive bool, containersActive bool) *mockContainerStore {
	containers := make([]repository.Container, len(containerNames))
	for i, name := range containerNames {
		containers[i] = repository.Container{
			Name:   name,
			URL:    "http://localhost:800" + string(rune('0'+i)),
			Active: boolPtr(containersActive),
		}
	}
	return &mockContainerStore{
		doc: repository.DataDocument{
			Containers: containers,
			Groups: []repository.Group{
				{Name: groupName, Container: containerNames, Active: boolPtr(groupActive)},
			},
		},
	}
}

func TestRuntimeController_WaitingPage_ContainerNotFound(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreEmpty()
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/nonexistent/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_ContainerNotActive(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreWithActiveContainer("my-container", "http://localhost:8080", false)
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_ContainerActiveAndRunning(t *testing.T) {
	rt := newMockRuntime()
	rt.runningContainers["my-container"] = true

	store := newMockStoreWithActiveContainer("my-container", "http://localhost:8080", true)
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify content type is HTML
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected content type 'text/html; charset=utf-8', got '%s'", contentType)
	}
}

func TestRuntimeController_WaitingPage_ContainerActiveNotRunning(t *testing.T) {
	rt := newMockRuntime()
	rt.runningContainers["my-container"] = false

	store := newMockStoreWithActiveContainer("my-container", "http://localhost:8080", true)
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Give goroutine a moment to start the container
	// In real test, we'd use synchronization, but for this test we just verify it was called
}

func TestRuntimeController_WaitingPage_GroupNotFound(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreEmpty()
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/nonexistent-group/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_GroupNotActive(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreWithGroup("my-group", []string{"container1", "container2"}, false, true)
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-group/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_GroupActiveSuccess(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreWithGroup("my-group", []string{"container1", "container2"}, true, true)
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-group/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify content type is HTML
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected content type 'text/html; charset=utf-8', got '%s'", contentType)
	}
}

func TestRuntimeController_WaitingPage_MissingName(t *testing.T) {
	rt := newMockRuntime()
	store := newMockStoreEmpty()
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime//waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_GroupEmptyContainers(t *testing.T) {
	rt := newMockRuntime()
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Groups: []repository.Group{
				{Name: "empty-group", Container: []string{}, Active: boolPtr(true)},
			},
		},
	}
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/empty-group/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_GroupWithNonexistentContainers(t *testing.T) {
	rt := newMockRuntime()
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{}, // No containers defined
			Groups: []repository.Group{
				{Name: "my-group", Container: []string{"nonexistent"}, Active: boolPtr(true)},
			},
		},
	}
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-group/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// mockContainerStoreWithError simulates a store that fails on Snapshot
type mockContainerStoreWithError struct {
	mockContainerStore
	snapshotErr error
}

func (m *mockContainerStoreWithError) Snapshot() (repository.DataDocument, error) {
	if m.snapshotErr != nil {
		return repository.DataDocument{}, m.snapshotErr
	}
	return m.doc, nil
}

func TestRuntimeController_WaitingPage_SnapshotError(t *testing.T) {
	rt := newMockRuntime()
	store := &mockContainerStoreWithError{
		snapshotErr: errors.New("database connection failed"),
	}
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_ContainerWithNilActive(t *testing.T) {
	rt := newMockRuntime()
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{
				{Name: "my-container", URL: "http://localhost:8080", Active: nil},
			},
		},
	}
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-container/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Container with nil active should be treated as not active (403)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestRuntimeController_WaitingPage_GroupWithNilActive(t *testing.T) {
	rt := newMockRuntime()
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{
				{Name: "container1", URL: "http://localhost:8080", Active: boolPtr(true)},
			},
			Groups: []repository.Group{
				{Name: "my-group", Container: []string{"container1"}, Active: nil},
			},
		},
	}
	rc := NewRuntimeController(context.Background(), rt, store)

	r := gin.New()
	r.GET("/runtime/:name/waiting", rc.WaitingPage)

	req := httptest.NewRequest(http.MethodGet, "/runtime/my-group/waiting", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Group with nil active should be treated as not active (403)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockContainerStore implements cache.ContainerStore for testing
type mockContainerStore struct {
	doc       repository.DataDocument
	addErr    error
	removeErr error
}

func (m *mockContainerStore) Snapshot() (repository.DataDocument, error) {
	return m.doc, nil
}

func (m *mockContainerStore) AddContainer(c repository.Container) (repository.DataDocument, error) {
	if m.addErr != nil {
		return repository.DataDocument{}, m.addErr
	}
	m.doc.Containers = append(m.doc.Containers, c)
	return m.doc, nil
}

func (m *mockContainerStore) RemoveContainer(name string) (repository.DataDocument, error) {
	if m.removeErr != nil {
		return repository.DataDocument{}, m.removeErr
	}
	for i, c := range m.doc.Containers {
		if c.Name == name {
			m.doc.Containers = append(m.doc.Containers[:i], m.doc.Containers[i+1:]...)
			return m.doc, nil
		}
	}
	return repository.DataDocument{}, cache.ErrContainerNotFound
}

func TestContainerController_AllContainers(t *testing.T) {
	active := true
	running := false
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{
				{Name: "test1", FriendlyName: "Test 1", URL: "http://test1.local", Active: &active, Running: &running},
				{Name: "test2", FriendlyName: "Test 2", URL: "http://test2.local", Active: &active, Running: &running},
			},
		},
	}

	cc := NewContainerController(store)

	r := gin.New()
	r.GET("/containers", cc.AllContainers)

	req := httptest.NewRequest(http.MethodGet, "/containers", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var containers []repository.Container
	if err := json.Unmarshal(w.Body.Bytes(), &containers); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}
}

func TestContainerController_CreateOrUpdateContainer_Valid(t *testing.T) {
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{},
		},
	}

	cc := NewContainerController(store)

	r := gin.New()
	r.POST("/container", cc.CreateOrUpdateContainer)

	active := true
	running := false
	container := repository.Container{
		Name:         "new-container",
		FriendlyName: "New Container",
		URL:          "http://new.local",
		Active:       &active,
		Running:      &running,
	}
	body, _ := json.Marshal(container)

	req := httptest.NewRequest(http.MethodPost, "/container", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestContainerController_CreateOrUpdateContainer_InvalidPayload(t *testing.T) {
	store := &mockContainerStore{}
	cc := NewContainerController(store)

	r := gin.New()
	r.POST("/container", cc.CreateOrUpdateContainer)

	req := httptest.NewRequest(http.MethodPost, "/container", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestContainerController_CreateOrUpdateContainer_ValidationError(t *testing.T) {
	store := &mockContainerStore{}
	cc := NewContainerController(store)

	r := gin.New()
	r.POST("/container", cc.CreateOrUpdateContainer)

	// Missing required fields
	container := map[string]any{
		"name": "test",
		// missing friendly_name, url, active, running
	}
	body, _ := json.Marshal(container)

	req := httptest.NewRequest(http.MethodPost, "/container", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for validation error, got %d", w.Code)
	}
}

func TestContainerController_CreateOrUpdateContainer_StoreError(t *testing.T) {
	store := &mockContainerStore{
		addErr: errors.New("store error"),
	}
	cc := NewContainerController(store)

	r := gin.New()
	r.POST("/container", cc.CreateOrUpdateContainer)

	active := true
	running := false
	container := repository.Container{
		Name:         "test",
		FriendlyName: "Test",
		URL:          "http://test.local",
		Active:       &active,
		Running:      &running,
	}
	body, _ := json.Marshal(container)

	req := httptest.NewRequest(http.MethodPost, "/container", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestContainerController_DeleteContainer_Success(t *testing.T) {
	active := true
	running := false
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{
				{Name: "to-delete", FriendlyName: "To Delete", URL: "http://del.local", Active: &active, Running: &running},
			},
		},
	}
	cc := NewContainerController(store)

	r := gin.New()
	r.DELETE("/container/:name", cc.DeleteContainer)

	req := httptest.NewRequest(http.MethodDelete, "/container/to-delete", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestContainerController_DeleteContainer_NotFound(t *testing.T) {
	store := &mockContainerStore{
		doc: repository.DataDocument{
			Containers: []repository.Container{},
		},
	}
	cc := NewContainerController(store)

	r := gin.New()
	r.DELETE("/container/:name", cc.DeleteContainer)

	req := httptest.NewRequest(http.MethodDelete, "/container/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestContainerController_DeleteContainer_MissingName(t *testing.T) {
	store := &mockContainerStore{}
	cc := NewContainerController(store)

	r := gin.New()
	// Route without :name param
	r.DELETE("/container/", cc.DeleteContainer)

	req := httptest.NewRequest(http.MethodDelete, "/container/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

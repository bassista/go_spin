package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func boolPtrJSON(b bool) *bool {
	return &b
}

func createTestDataDocument() DataDocument {
	return DataDocument{
		Metadata: Metadata{LastUpdate: 1000},
		Containers: []Container{
			{Name: "container1", FriendlyName: "Container 1", URL: "http://c1.local", Running: boolPtrJSON(false), Active: boolPtrJSON(true)},
		},
		Order: []string{"container1"},
		Groups: []Group{
			{Name: "group1", Container: []string{"container1"}, Active: boolPtrJSON(true)},
		},
		GroupOrder: []string{"group1"},
		Schedules: []Schedule{
			{ID: "schedule1", Target: "container1", TargetType: "container", Timers: []Timer{}},
		},
	}
}

func TestNewJSONRepository_Success(t *testing.T) {
	repo, err := NewJSONRepository("/tmp/test-config.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo == nil {
		t.Error("expected repository to be created")
	}
}

func TestNewJSONRepository_EmptyPath(t *testing.T) {
	_, err := NewJSONRepository("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestJSONRepository_LoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create initial file
	doc := createTestDataDocument()
	data, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, err := NewJSONRepository(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Load
	loaded, err := repo.Load()
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if len(loaded.Containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(loaded.Containers))
	}

	if loaded.Containers[0].Name != "container1" {
		t.Errorf("expected container name 'container1', got '%s'", loaded.Containers[0].Name)
	}
}

func TestJSONRepository_Load_FileNotFound(t *testing.T) {
	repo, _ := NewJSONRepository("/nonexistent/path/config.json")
	_, err := repo.Load()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestJSONRepository_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create invalid JSON file
	if err := os.WriteFile(configPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, _ := NewJSONRepository(configPath)
	_, err := repo.Load()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONRepository_Load_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create JSON with missing required fields
	invalidDoc := map[string]interface{}{
		"metadata": map[string]interface{}{"lastUpdate": 1000},
		"containers": []map[string]interface{}{
			{"name": "container1"}, // missing required fields
		},
	}
	data, _ := json.MarshalIndent(invalidDoc, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, _ := NewJSONRepository(configPath)
	_, err := repo.Load()
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestJSONRepository_Save_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create initial empty file
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, err := NewJSONRepository(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc := createTestDataDocument()
	err = repo.Save(&doc)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	var savedDoc DataDocument
	if err := json.Unmarshal(data, &savedDoc); err != nil {
		t.Fatalf("failed to parse saved file: %v", err)
	}

	if len(savedDoc.Containers) != 1 {
		t.Errorf("expected 1 container in saved file, got %d", len(savedDoc.Containers))
	}
}

func TestJSONRepository_Save_NilDocument(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	repo, _ := NewJSONRepository(configPath)
	err := repo.Save(nil)
	if err == nil {
		t.Error("expected error for nil document")
	}
}

func TestJSONRepository_Save_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	repo, _ := NewJSONRepository(configPath)

	// Document with invalid container (missing required URL)
	doc := DataDocument{
		Containers: []Container{
			{Name: "test", FriendlyName: "Test", URL: "", Running: boolPtrJSON(false), Active: boolPtrJSON(true)},
		},
	}

	err := repo.Save(&doc)
	if err == nil {
		t.Error("expected validation error")
	}
}

// MockCacheStore implements CacheStore for testing
type MockCacheStore struct {
	lastUpdate int64
	dirty      bool
	doc        DataDocument
	replaced   bool
}

func (m *MockCacheStore) GetLastUpdate() int64 {
	return m.lastUpdate
}

func (m *MockCacheStore) IsDirty() bool {
	return m.dirty
}

func (m *MockCacheStore) Snapshot() (DataDocument, error) {
	return m.doc, nil
}

func (m *MockCacheStore) Replace(doc DataDocument) error {
	m.doc = doc
	m.lastUpdate = doc.Metadata.LastUpdate
	m.replaced = true
	return nil
}

func TestJSONRepository_MakeWatcherCallback_ReloadsWhenDiskNewer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	doc := createTestDataDocument()
	doc.Metadata.LastUpdate = 2000
	data, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, _ := NewJSONRepository(configPath)
	jsonRepo := repo.(*JSONRepository)

	cache := &MockCacheStore{
		lastUpdate: 1000, // Older than disk
		dirty:      false,
		doc:        DataDocument{},
	}

	callback := jsonRepo.MakeWatcherCallback(cache)
	callback()

	if !cache.replaced {
		t.Error("expected cache to be replaced when disk is newer")
	}
}

func TestJSONRepository_MakeWatcherCallback_SkipsWhenDiskOlder(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	doc := createTestDataDocument()
	doc.Metadata.LastUpdate = 500
	data, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, _ := NewJSONRepository(configPath)
	jsonRepo := repo.(*JSONRepository)

	cache := &MockCacheStore{
		lastUpdate: 1000, // Newer than disk
		dirty:      false,
		doc:        DataDocument{},
	}

	callback := jsonRepo.MakeWatcherCallback(cache)
	callback()

	if cache.replaced {
		t.Error("expected cache NOT to be replaced when disk is older")
	}
}

func TestJSONRepository_MakeWatcherCallback_SkipsWhenDirty(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	doc := createTestDataDocument()
	doc.Metadata.LastUpdate = 2000
	data, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, _ := NewJSONRepository(configPath)
	jsonRepo := repo.(*JSONRepository)

	cache := &MockCacheStore{
		lastUpdate: 1000,
		dirty:      true, // Cache is dirty
		doc:        DataDocument{},
	}

	callback := jsonRepo.MakeWatcherCallback(cache)
	callback()

	if cache.replaced {
		t.Error("expected cache NOT to be replaced when dirty")
	}
}

func TestJSONRepository_MakeWatcherCallback_SkipsWhenSameContent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	doc := createTestDataDocument()
	doc.Metadata.LastUpdate = 1000
	data, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	repo, _ := NewJSONRepository(configPath)
	jsonRepo := repo.(*JSONRepository)

	cache := &MockCacheStore{
		lastUpdate: 1000, // Same as disk
		dirty:      false,
		doc:        doc, // Same content
	}

	callback := jsonRepo.MakeWatcherCallback(cache)
	callback()

	if cache.replaced {
		t.Error("expected cache NOT to be replaced when content is same")
	}
}

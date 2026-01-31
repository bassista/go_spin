package cache

import (
	"encoding/json"
	"sync"

	"github.com/bassista/go_spin/internal/repository"
)

// Store keeps an in-memory copy of the data document.
type Store struct {
	mu         sync.RWMutex
	data       repository.DataDocument
	dirty      bool  // true if cache changed since last persist
	lastUpdate int64 // cache's metadata.lastUpdate
}

// NewStore creates an empty cache store.
func NewStore(doc repository.DataDocument) *Store {
	return &Store{data: doc, lastUpdate: doc.Metadata.LastUpdate}
}

// MarkDirty sets the dirty flag to true.
func (s *Store) MarkDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = true
}

// IsDirty returns true if cache has uncommitted changes.
func (s *Store) IsDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

// ClearDirty resets the dirty flag.
func (s *Store) ClearDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = false
}

// GetLastUpdate returns the cache's last update timestamp.
func (s *Store) GetLastUpdate() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastUpdate
}

// SetLastUpdate sets the cache's last update timestamp.
func (s *Store) SetLastUpdate(ts int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastUpdate = ts
}

// Snapshot returns a deep copy of the cached data.
func (s *Store) Snapshot() (repository.DataDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneData(s.data)
}

// Replace swaps the cached data.
func (s *Store) Replace(doc repository.DataDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cloned, err := cloneData(doc)
	if err != nil {
		return err
	}
	s.data = cloned
	s.lastUpdate = doc.Metadata.LastUpdate
	s.dirty = false

	return nil
}

// AddContainer upserts a container by name, updating order and returning the new snapshot.
func (s *Store) AddContainer(container repository.Container) (repository.DataDocument, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clonedContainer, err := cloneContainer(container)
	if err != nil {
		return repository.DataDocument{}, err
	}

	inOrder := false
	for _, name := range s.data.Order {
		if name == clonedContainer.Name {
			inOrder = true
			break
		}
	}

	replaced := false
	for i := range s.data.Containers {
		if s.data.Containers[i].Name == clonedContainer.Name {
			s.data.Containers[i] = clonedContainer
			replaced = true
			break
		}
	}

	if !replaced {
		s.data.Containers = append(s.data.Containers, clonedContainer)
	}

	if !inOrder {
		s.data.Order = append(s.data.Order, clonedContainer.Name)
	}

	// Mark cache as dirty after mutation
	s.dirty = true

	return cloneData(s.data)
}

// cloneData deep-copies the document to avoid shared slices between cache and callers.
func cloneData(doc repository.DataDocument) (repository.DataDocument, error) {
	bytes, err := json.Marshal(doc)
	if err != nil {
		return repository.DataDocument{}, err
	}
	var copy repository.DataDocument
	if err := json.Unmarshal(bytes, &copy); err != nil {
		return repository.DataDocument{}, err
	}
	return copy, nil
}

// cloneContainer deep-copies a container to avoid shared pointer fields.
func cloneContainer(c repository.Container) (repository.Container, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return repository.Container{}, err
	}
	var copy repository.Container
	if err := json.Unmarshal(bytes, &copy); err != nil {
		return repository.Container{}, err
	}
	return copy, nil
}

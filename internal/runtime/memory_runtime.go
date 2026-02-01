package runtime

import (
	"context"
	"sync"

	"github.com/bassista/go_spin/internal/logger"
	"github.com/bassista/go_spin/internal/repository"
)

// MemoryRuntime is a temporary ContainerRuntime implementation that keeps state in memory.
// It is useful while the Docker-socket implementation is not available to execute tests or other development tasks.
type MemoryRuntime struct {
	mu      sync.RWMutex
	running map[string]bool
}

func NewMemoryRuntime() *MemoryRuntime {
	return &MemoryRuntime{running: map[string]bool{}}
}

func NewMemoryRuntimeFromDocument(doc repository.DataDocument) *MemoryRuntime {
	mr := NewMemoryRuntime()
	for _, c := range doc.Containers {
		if c.Name == "" {
			continue
		}
		if c.Running != nil {
			mr.running[c.Name] = *c.Running
		}
	}
	return mr
}

func (m *MemoryRuntime) IsRunning(_ context.Context, containerName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	isRunning := m.running[containerName]
	logger.WithComponent("memory-runtime").Debugf("checking if container is running: %s, result: %v", containerName, isRunning)
	return isRunning, nil
}

func (m *MemoryRuntime) Start(_ context.Context, containerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	logger.WithComponent("memory-runtime").Debugf("starting container: %s", containerName)
	m.running[containerName] = true
	return nil
}

func (m *MemoryRuntime) Stop(_ context.Context, containerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	logger.WithComponent("memory-runtime").Debugf("stopping container: %s", containerName)
	m.running[containerName] = false
	return nil
}

package runtime

import (
	"fmt"

	"github.com/bassista/go_spin/internal/repository"
)

const (
	RuntimeTypeDocker = "docker"
	RuntimeTypeMemory = "memory"
)

// NewRuntimeFromConfig creates a ContainerRuntime based on the runtime type.
// If runtimeType is "memory", it creates a MemoryRuntime initialized from the document.
// If runtimeType is "docker" (default), it creates a DockerRuntime.
func NewRuntimeFromConfig(runtimeType string, doc *repository.DataDocument) (ContainerRuntime, error) {
	switch runtimeType {
	case RuntimeTypeMemory:
		if doc != nil {
			return NewMemoryRuntimeFromDocument(*doc), nil
		}
		return NewMemoryRuntime(), nil
	case RuntimeTypeDocker, "":
		return NewDockerRuntime()
	default:
		return nil, fmt.Errorf("unknown runtime type: %s (supported: %s, %s)", runtimeType, RuntimeTypeDocker, RuntimeTypeMemory)
	}
}

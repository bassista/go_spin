package runtime

import "context"

// ContainerStats holds resource usage statistics for a container.
type ContainerStats struct {
	// CPUPercent is the percentage of CPU usage (0-100 per core, can exceed 100 on multi-core).
	CPUPercent float64
	// MemoryMB is the amount of memory used in megabytes.
	MemoryMB float64
}

// ContainerRuntime abstracts container lifecycle operations.
// A Docker-socket implementation will be added later.
type ContainerRuntime interface {
	IsRunning(ctx context.Context, containerName string) (bool, error)
	Start(ctx context.Context, containerName string) error
	Stop(ctx context.Context, containerName string) error
	// ListContainers returns the list of container names present in the runtime.
	// Names must be returned exactly as they are (case-sensitive).
	ListContainers(ctx context.Context) ([]string, error)
	// Stats returns CPU and memory usage statistics for a container.
	Stats(ctx context.Context, containerName string) (ContainerStats, error)
}

package runtime

import "context"

// ContainerRuntime abstracts container lifecycle operations.
// A Docker-socket implementation will be added later.
type ContainerRuntime interface {
	IsRunning(ctx context.Context, containerName string) (bool, error)
	Start(ctx context.Context, containerName string) error
	Stop(ctx context.Context, containerName string) error
	// ListContainers returns the list of container names present in the runtime.
	// Names must be returned exactly as they are (case-sensitive).
	ListContainers(ctx context.Context) ([]string, error)
}

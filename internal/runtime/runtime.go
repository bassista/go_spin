package runtime

import "context"

// ContainerRuntime abstracts container lifecycle operations.
// A Docker-socket implementation will be added later.
type ContainerRuntime interface {
	IsRunning(ctx context.Context, containerName string) (bool, error)
	Start(ctx context.Context, containerName string) error
	Stop(ctx context.Context, containerName string) error
}

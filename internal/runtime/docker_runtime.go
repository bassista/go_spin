package runtime

import (
	"context"
	"fmt"

	"github.com/bassista/go_spin/internal/logger"
	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

// DockerClient defines the interface for Docker client operations used by DockerRuntime.
// This interface allows for mocking in tests.
type DockerClient interface {
	ContainerInspect(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error)
	ContainerStart(ctx context.Context, containerID string, options client.ContainerStartOptions) (client.ContainerStartResult, error)
	ContainerStop(ctx context.Context, containerID string, options client.ContainerStopOptions) (client.ContainerStopResult, error)
}

type DockerRuntime struct {
	cli DockerClient
}

func NewDockerRuntime() (*DockerRuntime, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}
	return &DockerRuntime{cli: cli}, nil
}

// NewDockerRuntimeWithClient creates a DockerRuntime with a custom client.
// This is primarily used for testing purposes.
func NewDockerRuntimeWithClient(cli DockerClient) *DockerRuntime {
	return &DockerRuntime{cli: cli}
}

func (d *DockerRuntime) IsRunning(ctx context.Context, containerName string) (bool, error) {
	logger.WithComponent("docker").Debugf("checking if container is running: %s", containerName)
	inspect, err := d.cli.ContainerInspect(ctx, containerName, client.ContainerInspectOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			logger.WithComponent("docker").Debugf("container not found: %s", containerName)
			return false, fmt.Errorf("container %s not found", containerName)
		}
		logger.WithComponent("docker").Errorf("failed to inspect container %s: %v", containerName, err)
		return false, fmt.Errorf("error checking status of container %s: %w", containerName, err)
	}

	if inspect.Container.State == nil {
		logger.WithComponent("docker").Warnf("container state is null: %s", containerName)
		return false, nil
	}
	logger.WithComponent("docker").Debugf("container isRunning %t for : %s", inspect.Container.State.Running, containerName)
	logger.WithComponent("docker").Debugf("container status %s for : %s", inspect.Container.State.Status, containerName)
	return inspect.Container.State.Running, nil
}

func (d *DockerRuntime) Start(ctx context.Context, containerName string) error {
	logger.WithComponent("docker").Debugf("starting container: %s", containerName)
	_, err := d.cli.ContainerStart(ctx, containerName, client.ContainerStartOptions{})
	if err != nil {
		logger.WithComponent("docker").Errorf("failed to start container %s: %v", containerName, err)
		return fmt.Errorf("error starting container %s: %w", containerName, err)
	}
	logger.WithComponent("docker").Debugf("container started successfully: %s", containerName)
	return nil
}

func (d *DockerRuntime) Stop(ctx context.Context, containerName string) error {
	logger.WithComponent("docker").Debugf("stopping container: %s", containerName)
	_, err := d.cli.ContainerStop(ctx, containerName, client.ContainerStopOptions{})
	if err != nil {
		logger.WithComponent("docker").Errorf("failed to stop container %s: %v", containerName, err)
		return fmt.Errorf("error stopping container %s: %w", containerName, err)
	}
	logger.WithComponent("docker").Debugf("container stopped successfully: %s", containerName)
	return nil
}

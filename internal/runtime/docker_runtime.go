package runtime

import (
	"context"
	"fmt"

	"github.com/bassista/go_spin/internal/logger"
	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

type DockerRuntime struct {
	cli *client.Client
}

func NewDockerRuntime() (*DockerRuntime, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}
	return &DockerRuntime{cli: cli}, nil
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
		return false, nil
	}
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

package runtime

import (
	"context"
	"fmt"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

type DockerRuntime struct {
	cli *client.Client
}

func NewDockerRuntime() (*DockerRuntime, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("errore nella creazione del client Docker: %w", err)
	}
	return &DockerRuntime{cli: cli}, nil
}

func (d *DockerRuntime) IsRunning(ctx context.Context, containerName string) (bool, error) {
	inspect, err := d.cli.ContainerInspect(ctx, containerName, client.ContainerInspectOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, fmt.Errorf("container %s non trovato", containerName)
		}
		return false, fmt.Errorf("errore nel controllo dello stato del container %s: %w", containerName, err)
	}

	if inspect.Container.State == nil {
		return false, nil
	}
	return inspect.Container.State.Running, nil
}

func (d *DockerRuntime) Start(ctx context.Context, containerName string) error {
	_, err := d.cli.ContainerStart(ctx, containerName, client.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("errore nell'avvio del container %s: %w", containerName, err)
	}
	return nil
}

func (d *DockerRuntime) Stop(ctx context.Context, containerName string) error {
	_, err := d.cli.ContainerStop(ctx, containerName, client.ContainerStopOptions{})
	if err != nil {
		return fmt.Errorf("errore nell'arresto del container %s: %w", containerName, err)
	}
	return nil
}

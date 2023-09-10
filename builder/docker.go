package builder

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
)

// Docker is a docker image builder.
type Docker struct {
	cli    *client.Client
	logger *slog.Logger
}

// NewDocker returns new Docker instance.
func NewDocker(logger *slog.Logger, cli *client.Client) *Docker {
	return &Docker{cli: cli, logger: logger}
}

// ImageBuild builds docker image.
func (d Docker) ImageBuild(ctx context.Context, dst, name string) (string, error) {
	tar, err := archive.TarWithOptions(dst, &archive.TarOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create tar: %w", err)
	}

	tag := "go-lambda:" + name

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
		Remove:     true,
	}

	res, err := d.cli.ImageBuild(ctx, tar, opts)
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}

	defer res.Body.Close()

	if err := checkErr(res.Body); err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}

	d.logger.Debug("image built", slog.String("tag", tag))

	return tag, nil
}

// ContainerCreate creates Docker container.
func (d Docker) ContainerCreate(ctx context.Context, image string, port int) (string, error) {
	resp, err := d.cli.ContainerCreate(
		ctx, &container.Config{
			Image: image,
			Cmd:   []string{},
			Tty:   false,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: strconv.Itoa(port),
					},
				},
			},
		},
		nil,
		nil,
		"go-lambda",
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	d.logger.Info("container created", slog.String("id", resp.ID[:5]), slog.Int("port", port))

	return resp.ID, nil
}

// ContainerStart starts Docker container.
func (d Docker) ContainerStart(ctx context.Context, imageID string) error {
	if err := d.cli.ContainerStart(ctx, imageID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Debug("container started", slog.String("id", imageID[:5]))

	return nil
}

// ContainerStop stops Docker container.
func (d Docker) ContainerStop(ctx context.Context, imageID string) error {
	if err := d.cli.ContainerStop(ctx, imageID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	d.logger.Debug("container stopped", slog.String("id", imageID[:5]))

	return nil
}

// ContainersList lists all Docker containers.
func (d Docker) ContainersList(ctx context.Context) ([]types.Container, error) {
	containers, err := d.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, nil
}

// ContainerInspect inspects Docker container.
func (d Docker) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	data, err := d.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return types.ContainerJSON{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	return data, nil
}

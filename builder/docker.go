package builder

import (
	"context"
	"fmt"
	"log/slog"

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

// Build builds docker image.
func (d Docker) Build(ctx context.Context, file string) (string, error) {
	tar, err := archive.TarWithOptions("infra/", &archive.TarOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create tar: %w", err)
	}

	tag := "ihippik/go-lambda:v1.0.0"

	opts := types.ImageBuildOptions{
		Dockerfile: file,
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

// Run create and runs docker container.
func (d Docker) Run(ctx context.Context, image string) error {
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
						HostPort: "8080",
					},
				},
			},
		},
		nil,
		nil,
		"go-lambda",
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	d.logger.Debug("container created", slog.String("id", resp.ID))

	if err := d.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Debug("container started", slog.String("id", resp.ID))

	return nil
}

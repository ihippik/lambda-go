package lambda

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ihippik/lambda-go/config"
)

type builder interface {
	ImageBuild(ctx context.Context, file string, name string) (string, error)
	ContainerCreate(ctx context.Context, image string, port int) (string, error)
	ContainerStart(ctx context.Context, containerID string) error
	ContainerStop(ctx context.Context, containerID string) error
}

// Service is a service for lambda.
type Service struct {
	cfg      *config.Config
	log      *slog.Logger
	client   *http.Client
	builder  builder
	register sync.Map
}

// NewService returns new Service instance.
func NewService(cfg *config.Config, log *slog.Logger, container builder) *Service {
	return &Service{
		cfg:     cfg,
		log:     log,
		builder: container,
		client:  http.DefaultClient,
	}
}

func (s *Service) Create(ctx context.Context, name string, file io.ReadCloser) error {
	port := rand.Intn(65535-1024) + 1024

	if err := s.decompress("infra", file); err != nil {
		return fmt.Errorf("decompress: %w", err)
	}

	img, err := s.builder.ImageBuild(ctx, "infra", name)
	if err != nil {
		return fmt.Errorf("build image: %w", err)
	}

	s.log.Info("build image", "image", img)

	containerID, err := s.builder.ContainerCreate(ctx, img, port)
	if err != nil {
		return fmt.Errorf("run builder: %w", err)
	}

	s.register.Store(name, newMetaData(containerID, port))

	return nil
}

func (s *Service) Invoke(ctx context.Context, name string, data []byte) ([]byte, error) {
	value, ok := s.register.Load(name)
	if !ok {
		return nil, fmt.Errorf("function %s not found", name)
	}

	containerMeta, ok := value.(*metaData)
	if !ok {
		return nil, errors.New("invalid container meta type")
	}

	if err := s.builder.ContainerStart(ctx, containerMeta.containerID); err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	// FIXME: need to wait for container start

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		containerMeta.address(),
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("post request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if err := s.builder.ContainerStop(ctx, containerMeta.containerID); err != nil {
		return nil, fmt.Errorf("stop container: %w", err)
	}

	return respData, nil
}

func (s *Service) decompress(dst string, file io.ReadCloser) error {
	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header

	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		if strings.HasPrefix(header.Name, ".") {
			continue // TODO: need to Google it :)
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(target, 0755); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
			s.log.Debug("create dir", "path", target)
		case tar.TypeReg:
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}

			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			s.log.Debug("create file", "path", target)
		default:
			return errors.New("unknown type")
		}
	}

	return nil
}

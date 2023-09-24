package lambda

import (
	"archive/tar"
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
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ihippik/lambda-go/config"
	"github.com/ihippik/lambda-go/lambda/proto"
)

type builder interface {
	ImageBuild(ctx context.Context, file string, name string) (string, error)
	ContainerCreate(ctx context.Context, image string, port int) (string, error)
	ContainerStart(ctx context.Context, containerID string) error
	ContainerStop(ctx context.Context, containerID string) error
	ContainersList(ctx context.Context) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
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

// Init initializes service.
// It gets all containers with name "go-lambda" and registers them in the service.
func (s *Service) Init(ctx context.Context) error {
	containers, err := s.builder.ContainersList(ctx)
	if err != nil {
		return fmt.Errorf("list containers: %w", err)
	}

	for _, container := range containers {
		if len(container.Names) > 0 && container.Names[0] != "/go-lambda" {
			continue
		}

		data, err := s.builder.ContainerInspect(ctx, container.ID)
		if err != nil {
			return fmt.Errorf("inspect container: %w", err)
		}

		funcName, port, err := s.parseContainerData(data)
		if err != nil {
			return fmt.Errorf("parse container data: %w", err)
		}

		s.register.Store(funcName, newMetaData(container.ID, port))

		s.log.Info("init: register function", "name", funcName, "port", port)
	}

	return nil
}

// Create creates new lambda function. If function with the same name already exists, it will skip.
func (s *Service) Create(ctx context.Context, name string, file io.ReadCloser) error {
	// TODO: if container exists, need to overwrite it
	if _, exists := s.register.Load(name); exists {
		s.log.Info("container already exists", slog.String("name", name))
		return nil
	}

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

	respData, err := s.makeRequest(ctx, data, containerMeta)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}

	if err := s.builder.ContainerStop(ctx, containerMeta.containerID); err != nil {
		return nil, fmt.Errorf("stop container: %w", err)
	}

	return respData, nil
}

// makeRequest makes http request to container with Lambda.
// Using retry pattern for waiting container ready for requests.
func (s *Service) makeRequest(ctx context.Context, data []byte, meta *metaData) ([]byte, error) {
	s.log.Info("make request", "size", len(data), "address", meta.address())

	conn, err := grpc.Dial(meta.address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create a gRPC client.
	client := proto.NewLambdaServerClient(conn)

	resp, err := client.MakeRequest(ctx, &proto.Payload{
		Data: data,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// decompress decompresses tar.gz archive.
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

// parseContainerData parses container data and returns container image tag (user func name) and port.
func (s *Service) parseContainerData(data types.ContainerJSON) (string, int, error) {
	image := strings.Split(data.Config.Image, ":")
	if len(image) != 2 {
		return "", 0, errors.New("invalid image name")
	}

	tag := image[1]

	for _, v := range data.ContainerJSONBase.HostConfig.PortBindings {
		if len(v) == 0 {
			continue
		}

		port, err := strconv.Atoi(v[0].HostPort)
		if err != nil {
			return "", 0, fmt.Errorf("parse port: %w", err)
		}

		return tag, port, nil
	}

	return "", 0, errors.New("port not found")
}

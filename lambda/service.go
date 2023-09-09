package lambda

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ihippik/lambda-go/config"
)

type container interface {
	Build(ctx context.Context, file string) (string, error)
	Run(ctx context.Context, image string) error
}

// Service is a service for lambda.
type Service struct {
	cfg       *config.Config
	log       *slog.Logger
	container container
	register  map[string]struct{}
}

// NewService returns new Service instance.
func NewService(cfg *config.Config, log *slog.Logger, container container) *Service {
	return &Service{
		cfg:       cfg,
		log:       log,
		container: container,
		register:  make(map[string]struct{}),
	}
}

func (s *Service) Create(ctx context.Context, name string) error {
	img, err := s.container.Build(ctx, "Dockerfile")
	if err != nil {
		return fmt.Errorf("build image: %w", err)
	}

	s.log.Info("build image", "image", img)

	if err := s.container.Run(ctx, img); err != nil {
		return fmt.Errorf("run container: %w", err)
	}

	s.register[name] = struct{}{}

	return nil
}

func (s *Service) Invoke(ctx context.Context, name string, data []byte) ([]byte, error) {
	if _, ok := s.register[name]; !ok {
		return nil, fmt.Errorf("function %s not found", name)
	}

	resp, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://localhost"+s.cfg.App.FuncAddr,
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("post request: %w", err)
	}

	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return respData, nil
}

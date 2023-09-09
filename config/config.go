package config

import (
	"context"
	"fmt"

	cfg "github.com/ihippik/config"
	"github.com/sethvargo/go-envconfig"
)

// Config is a configuration for the service.
type Config struct {
	Log *cfg.Logger `env:",prefix=LOG_,required"`
	App AppCfg      `env:",prefix=APP_,required"`
}

type AppCfg struct {
	FuncAddr   string `env:"FUNC_ADDR,required"`
	ServerAddr string `env:"SERVER_ADDR,required"`
}

// NewConfig returns new Config.
func NewConfig(ctx context.Context) (*Config, error) {
	var conf Config

	if err := envconfig.Process(ctx, &conf); err != nil {
		return nil, fmt.Errorf("process env: %w", err)
	}

	return &conf, nil
}

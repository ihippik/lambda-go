package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/docker/docker/client"
	cfg "github.com/ihippik/config"

	"github.com/ihippik/lambda-go/builder"
	"github.com/ihippik/lambda-go/config"
	"github.com/ihippik/lambda-go/lambda"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conf, err := config.NewConfig(ctx)
	if err != nil {
		slog.Error("new config", "err", err)
		return
	}

	logger := cfg.InitSlog(conf.Log, cfg.GetVersion(), true)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		slog.Error("new client", "err", err)
		return
	}

	c := builder.NewDocker(logger, cli)
	svc := lambda.NewService(conf, logger, c)
	edp := lambda.NewEndpoint(svc, logger, conf.App.ServerAddr)

	if err := edp.StartServer(ctx); err != nil {
		slog.Error("run", "err", err)
	}
}

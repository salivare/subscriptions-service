package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/app"
	"github.com/salivare/subscriptions-service/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	application, err := app.New(log, cfg)
	if err != nil {
		log.Error("failed to init application", slog.String("err", err.Error()))
		os.Exit(1)
	}

	go application.HTTPSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("Shutting down...")

	application.HTTPSrv.Stop()

	log.Info("Goodbye!")
}

func setupLogger(env string) *slogx.Logger {
	var level slog.Level

	switch env {
	case envLocal:
		level = slogx.LevelTrace
	case envDev:
		level = slog.LevelDebug
	case envProd:
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo
	}

	return slogx.New(
		slogx.WithLevel(level),
		slogx.WithContextKeys("trace_id", "request_id"),
		slogx.WithRemoval(slogx.NewRemovalSet().Add("bearer_token")),
	)
}

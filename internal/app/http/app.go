package httpapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/config"
)

type App struct {
	log             *slogx.Logger
	server          *http.Server
	host            string
	port            int
	shutdownTimeout time.Duration
}

func New(
	log *slogx.Logger,
	cfg config.HTTPConfig,
	router http.Handler,
) *App {
	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler:           router,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	return &App{
		log:             log,
		server:          srv,
		host:            cfg.Host,
		port:            cfg.Port,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	addr := net.JoinHostPort(a.host, strconv.Itoa(a.port))
	l, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log := a.log.With(
		slog.String("op", op),
		slog.String("addr", l.Addr().String()),
	)

	log.Info("HTTP server is starting")

	if err := a.server.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "httpapp.Stop"

	log := a.log.With(
		slog.String("op", op),
		slog.String("host", a.host),
		slog.Int("port", a.port),
	)

	log.Info("HTTP server is stopping")

	done := make(chan struct{})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
		defer cancel()

		if err := a.server.Shutdown(ctx); err != nil {
			log.Error("HTTP server shutdown error", slogx.Err(err))
		}
		close(done)
	}()

	select {
	case <-done:
		log.Info("HTTP server stopped gracefully")
	case <-time.After(a.shutdownTimeout):
		log.Warn("graceful stop timed out, forcing close")
		if err := a.server.Close(); err != nil {
			log.Error("HTTP server close error", slogx.Err(err))
		}
	}
}

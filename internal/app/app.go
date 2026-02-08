package app

import (
	"github.com/salivare-io/slogx"
	httpapp "github.com/salivare/subscriptions-service/internal/app/http"
	swaggerapp "github.com/salivare/subscriptions-service/internal/app/swagger"
	"github.com/salivare/subscriptions-service/internal/config"
	"github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1/save"
	"github.com/salivare/subscriptions-service/internal/httpserver/middleware"
	"github.com/salivare/subscriptions-service/internal/httpserver/router"
	"github.com/salivare/subscriptions-service/internal/storage/postgres"
)

// App is a root structure that aggregates all application modules
type App struct {
	HTTPSrv *httpapp.App
}

// New creates a new instance of the root application.
func New(log *slogx.Logger, cfg *config.Config) (*App, error) {
	r := router.New()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(middleware.LoggerContext(log))

	storage, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Error("could not connect to postgres", slogx.Err(err))
		return nil, err
	}

	r.POST("/api/v1/subscription", savev1.New(storage))

	sw := swaggerapp.New(
		cfg.SwaggerServer.JSONPath,
		cfg.SwaggerServer.UIPath,
	)
	sw.Register(r.Mux())

	httpApp := httpapp.New(log, cfg.HTTPServer, r)

	return &App{
		HTTPSrv: httpApp,
	}, nil
}

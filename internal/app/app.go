package app

import (
	"github.com/salivare-io/slogx"
	httpapp "github.com/salivare/subscriptions-service/internal/app/http"
	swaggerapp "github.com/salivare/subscriptions-service/internal/app/swagger"
	"github.com/salivare/subscriptions-service/internal/config"
	deletev1 "github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1/delete"
	getv1 "github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1/get"
	savev1 "github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1/save"
	sumv1 "github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1/sum"
	updatev1 "github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1/update"
	"github.com/salivare/subscriptions-service/internal/httpserver/middleware"
	"github.com/salivare/subscriptions-service/internal/httpserver/router"
	"github.com/salivare/subscriptions-service/internal/services/subscription"
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

	subSrv := subscription.New(storage, storage, storage, storage, storage)

	r.POST("/api/v1/subscription", savev1.New(subSrv))
	r.DELETE("/api/v1/subscription/{id}", deletev1.New(subSrv))
	r.PATCH("/api/v1/subscription/{id}", updatev1.New(subSrv))
	r.GET("/api/v1/subscription/{id}", getv1.New(subSrv))
	r.POST("/api/v1/subscription/sum", sumv1.New(subSrv))

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

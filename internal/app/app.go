package app

import (
	"github.com/salivare-io/slogx"
	httpapp "github.com/salivare/subscriptions-service/internal/app/http"
	swaggerapp "github.com/salivare/subscriptions-service/internal/app/swagger"
	"github.com/salivare/subscriptions-service/internal/config"
	"github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1/save"
	"github.com/salivare/subscriptions-service/internal/httpserver/middleware"
	"github.com/salivare/subscriptions-service/internal/httpserver/router"
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

	// заглушка
	storage := &stubSubscriptionSaver{}

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

type stubSubscriptionSaver struct{}

func (s *stubSubscriptionSaver) SaveSubscription(subscription savev1.Subscription) (int64, error) {
	return 1, nil
}

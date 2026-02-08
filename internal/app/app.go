package app

import (
	"github.com/salivare-io/slogx"
	httpapp "github.com/salivare/subscriptions-service/internal/app/http"
	"github.com/salivare/subscriptions-service/internal/config"
	"github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/save"
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

	r.POST("/subscription", save.New(storage))

	httpApp := httpapp.New(log, cfg.HTTPServer, r)

	return &App{
		HTTPSrv: httpApp,
	}, nil
}

type stubSubscriptionSaver struct{}

func (s *stubSubscriptionSaver) SaveSubscription(subscription save.Subscription) (int64, error) {
	return 1, nil
}

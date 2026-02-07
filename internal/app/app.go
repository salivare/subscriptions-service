package app

import (
	"net/http"

	"github.com/salivare-io/slogx"
	httpapp "github.com/salivare/subscriptions-service/internal/app/http"
	"github.com/salivare/subscriptions-service/internal/config"
)

type App struct {
	HTTPSrv *httpapp.App
}

func New(log *slogx.Logger, cfg *config.Config) (*App, error) {
	httpApp := httpapp.New(log, cfg.HTTPServer, http.NotFoundHandler())

	return &App{
		HTTPSrv: httpApp,
	}, nil
}

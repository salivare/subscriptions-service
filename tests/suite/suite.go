package suite

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/salivare/subscriptions-service/internal/config"
)

type Suite struct {
	*testing.T
	Cfg    *config.Config
	Client *http.Client
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./configs/testlocal.yaml"
	}

	cfg := config.MustLoadByPath(path)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	client := &http.Client{
		Timeout: 25 * time.Second,
	}

	return ctx, &Suite{
		T:      t,
		Cfg:    cfg,
		Client: client,
	}
}

func (s *Suite) URL(path string) string {
	return "http://" + net.JoinHostPort(s.Cfg.HTTPServer.Host, strconv.Itoa(s.Cfg.HTTPServer.Port)) + path
}

package render

import (
	"encoding/json"
	"net/http"

	"github.com/salivare/subscriptions-service/internal/httpserver/middleware"
)

type StatusCoder interface {
	StatusCode() int
}

func JSON(w http.ResponseWriter, r *http.Request, v any) {
	w.Header().Set("Content-Type", "application/json")

	if reqID := middleware.GetRequestID(r.Context()); reqID != "" {
		w.Header().Set("X-Request-ID", reqID)
	}

	if sc, ok := v.(StatusCoder); ok {
		w.WriteHeader(sc.StatusCode())
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(v)
}

func Bind(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

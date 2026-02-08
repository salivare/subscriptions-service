package render

import (
	"encoding/json"
	"net/http"

	"github.com/salivare/subscriptions-service/internal/httpserver/middleware"
	"github.com/salivare/subscriptions-service/internal/lib/api/response"
)

func JSON(w http.ResponseWriter, r *http.Request, v any) {
	w.Header().Set("Content-Type", "application/json")

	switch resp := v.(type) {
	case response.Response:
		if resp.Status == response.StatusError {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	default:
		w.WriteHeader(http.StatusOK)
	}

	if reqID := middleware.GetRequestID(r.Context()); reqID != "" {
		w.Header().Set("X-Request-ID", reqID)
	}

	_ = json.NewEncoder(w).Encode(v)
}

func Bind(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

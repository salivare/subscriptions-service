package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
	"github.com/salivare/subscriptions-service/internal/httpserver/router"
)

func ExtractID(w http.ResponseWriter, r *http.Request, log *slogx.Logger) (uuid.UUID, bool) {
	idStr := router.PathValue(r, "id")
	if idStr == "" {
		log.ErrorContext(r.Context(), "missing id in path")
		render.JSON(w, r, response.Error("id is required"))
		return uuid.UUID{}, false
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.ErrorContext(r.Context(), "invalid uuid", slogx.Err(err))
		render.JSON(w, r, response.Error("invalid id"))
		return uuid.UUID{}, false
	}

	return id, true
}

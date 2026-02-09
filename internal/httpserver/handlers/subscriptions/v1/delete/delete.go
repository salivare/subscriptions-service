package deletev1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
	"github.com/salivare/subscriptions-service/internal/httpserver/router"
	subSrv "github.com/salivare/subscriptions-service/internal/services/subscription"
)

type Subscription interface {
	Delete(ctx context.Context, id uuid.UUID) error
}

// New creates a handler for delete a subscription.
//
// @Summary      Delete subscription
// @Description  Deletes a subscription by ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subscription ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response  "Invalid ID"
// @Failure      404  {object}  response.Response  "Subscription not found"
// @Failure      500  {object}  response.Response  "Internal server error"
// @Router       /api/v1/subscription/{id} [delete]
func New(subscription Subscription) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.subscriptions.delete.New"
		ctx := r.Context()
		log := slogx.FromContext(ctx).With(slog.String("op", op))

		idStr := router.PathValue(r, "id")
		if idStr == "" {
			log.ErrorContext(ctx, "missing id in path")
			render.JSON(
				w, r, response.Error("id is required"),
			)
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			log.ErrorContext(ctx, "invalid uuid", slogx.Err(err))
			render.JSON(w, r, response.Error("invalid id"))
			return
		}

		err = subscription.Delete(ctx, id)
		if err != nil {
			if errors.Is(err, subSrv.ErrNotFound) {
				log.ErrorContext(ctx, "subscription not found", slogx.Err(err))
				render.JSON(
					w, r, response.Response{
						Status: response.StatusError,
						Error:  response.ErrNotFound,
						Code:   http.StatusNotFound,
					},
				)
				return
			}

			log.ErrorContext(ctx, "failed to delete subscription", slogx.Err(err))
			render.JSON(w, r, response.Internal("internal error"))
			return
		}

		render.JSON(w, r, response.OK())
	}
}

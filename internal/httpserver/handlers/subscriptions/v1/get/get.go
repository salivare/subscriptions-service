package getv1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	v1 "github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
	subSrv "github.com/salivare/subscriptions-service/internal/services/subscription"
)

type Subscription interface {
	Get(ctx context.Context, id uuid.UUID) (models.Subscription, error)
}

// New creates a handler for get a subscription.
//
// @Summary      Get subscription
// @Description  Get subscription by ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string              true  "Subscription ID (UUID)"
// @Success      200  {object}  response.Response   "Subscription data"
// @Failure      400  {object}  response.Response   "Invalid ID"
// @Failure      404  {object}  response.Response   "Subscription not found"
// @Failure      500  {object}  response.Response   "Internal error"
// @Router       /api/v1/subscription/{id} [get]
func New(subscription Subscription) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.subscriptions.get.New"
		ctx := r.Context()
		log := slogx.FromContext(ctx).With(slog.String("op", op))

		id, ok := v1.ExtractID(w, r, log)
		if !ok {
			return
		}

		sub, err := subscription.Get(ctx, id)

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

			log.ErrorContext(ctx, "failed to get subscription", slogx.Err(err))
			render.JSON(w, r, response.Internal("internal error"))
			return
		}

		render.JSON(w, r, response.OKWithData(sub))
	}
}

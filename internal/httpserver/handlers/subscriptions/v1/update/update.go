package updatev1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	v1 "github.com/salivare/subscriptions-service/internal/httpserver/handlers/subscriptions/v1"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/request"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
	subSrv "github.com/salivare/subscriptions-service/internal/services/subscription"
)

type Subscription interface {
	Update(ctx context.Context, id uuid.UUID, updateReq request.UpdateRequest) (models.Subscription, error)
}

// New creates a handler for update a subscription.
//
//	@Summary		Update subscription
//	@Description	Partially update subscription fields (PATCH). Any field may be omitted.
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Subscription ID (UUID)"
//	@Param			body	body		request.UpdateRequest			true	"Fields to update"
//	@Success		200		{object}	response.SubscriptionResponse	"Updated subscription"
//	@Failure		400		{object}	response.Response				"Invalid input"
//	@Failure		404		{object}	response.Response				"Subscription not found"
//	@Failure		500		{object}	response.Response				"Internal error"
//	@Router			/api/v1/subscription/{id} [patch]
func New(subscription Subscription) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.subscriptions.update.New"
		ctx := r.Context()
		log := slogx.FromContext(ctx).With(slog.String("op", op))

		id, ok := v1.ExtractID(w, r, log)
		if !ok {
			return
		}

		var reqBody request.UpdateRequest
		if err := render.Bind(r, &reqBody); err != nil {
			log.ErrorContext(ctx, "invalid json", slogx.Err(err))
			render.JSON(w, r, response.Error("invalid json"))
			return
		}

		if err := validator.New().Struct(reqBody); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.ErrorContext(ctx, "invalid request", slogx.Err(err))
			render.JSON(w, r, request.ValidationError(validateErr))
			return
		}

		updated, err := subscription.Update(ctx, id, reqBody)
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

			log.ErrorContext(ctx, "failed to update subscription", slogx.Err(err))
			render.JSON(w, r, response.Internal("internal error"))
			return
		}

		render.JSON(
			w, r, response.Response{
				Status: response.StatusOK,
				Data:   response.ToSubscriptionResponse(updated),
			},
		)
	}
}

package savev1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/request"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
	"github.com/salivare/subscriptions-service/internal/storage"
)

type Response struct {
	response.Response
	ID *uuid.UUID `json:"id,omitempty"`
}

func (r Response) StatusCode() int {
	return r.Response.StatusCode()
}

type SubscriptionSaver interface {
	SaveSubscription(ctx context.Context, subscription models.Subscription) (uuid.UUID, error)
}

// New creates a handler for creating a subscription.
//
// @Summary Create subscription
// @Description Creates a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body request.Request true "Subscription data"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /api/v1/subscription [post]
func New(subscriptionSaver SubscriptionSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.subscriptions.save.New"
		ctx := r.Context()
		log := slogx.FromContext(ctx).With(slog.String("op", op))

		var req request.Request
		if err := render.Bind(r, &req); err != nil {
			log.ErrorContext(ctx, "invalid json", slogx.Err(err))
			render.JSON(
				w, r, Response{
					Response: response.Error("invalid json"),
				},
			)
			return
		}

		if err := validator.New().Struct(req); err != nil {
			var validateError validator.ValidationErrors
			errors.As(err, &validateError)

			log.ErrorContext(ctx, "invalid request", slogx.Err(err))
			render.JSON(w, r, response.ValidationError(validateError))
			return
		}

		sub, err := req.ToModel()
		if err != nil {
			log.ErrorContext(ctx, "failed to convert request to model", slogx.Err(err))
			render.JSON(
				w, r, Response{
					Response: response.Error("invalid subscription data"),
				},
			)
			return
		}

		id, err := subscriptionSaver.SaveSubscription(ctx, sub)
		if err != nil {
			if errors.Is(err, storage.ErrSubscriptionExists) {
				log.WarnContext(ctx, "subscription already exists", slogx.Err(err))
				render.JSON(
					w, r, Response{
						Response: response.Conflict("subscription already exists"),
					},
				)
				return
			}

			log.ErrorContext(ctx, "failed to save subscription", slogx.Err(err))
			render.JSON(
				w, r, Response{
					Response: response.Internal("internal error"),
				},
			)
			return
		}

		render.JSON(
			w, r, Response{
				Response: response.OK(),
				ID:       &id,
			},
		)
	}
}

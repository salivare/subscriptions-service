package savev1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/request"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
	subSrv "github.com/salivare/subscriptions-service/internal/services/subscription"
)

type Response struct {
	response.Response
}

type CreateResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
}

func (r Response) StatusCode() int {
	return r.Response.StatusCode()
}

// Subscription service interface
type Subscription interface {
	Save(ctx context.Context, subscription models.Subscription) (uuid.UUID, time.Time, error)
}

// New creates a handler for creating a subscription.
//
//	@Summary		Create subscription
//	@Description	Creates a new subscription for a user
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.CreateRequest	true	"Subscription data"
//	@Success		200		{object}	CreateResponse
//	@Failure		400		{object}	Response	"Invalid request"
//	@Failure		409		{object}	Response	"Subscription already exists"
//	@Failure		500		{object}	Response	"Internal server error"
//	@Router			/api/v1/subscription [post]
func New(subscription Subscription) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.subscriptions.save.New"
		ctx := r.Context()
		log := slogx.FromContext(ctx).With(slog.String("op", op))

		var req request.CreateRequest
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
			render.JSON(w, r, request.ValidationError(validateError))
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

		id, createAt, err := subscription.Save(ctx, sub)
		if err != nil {
			if errors.Is(err, subSrv.ErrAlreadyExists) {
				render.JSON(
					w, r, Response{
						Response: response.Conflict(err.Error()),
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
			w, r, response.Response{
				Status: response.StatusOK,
				Data: CreateResponse{
					ID:        id,
					CreatedAt: createAt.Format(time.DateTime),
				},
			},
		)
	}
}

package sumv1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/request"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
)

type Subscription interface {
	Sum(ctx context.Context, f models.SumFilter) (int64, error)
}

// New creates a handler for calculating total subscription cost.
//
//	@Summary		Calculate total subscription cost
//	@Description	Sum of subscriptions for selected periods with optional filters
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.SumRequest	true	"Filters"
//	@Success		200		{object}	response.SumResponse
//	@Failure		400		{object}	response.Response	"Invalid request"
//	@Failure		500		{object}	response.Response	"Internal error"
//	@Router			/api/v1/subscription/sum [post]
func New(subscription Subscription) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.subscriptions.sum.New"
		ctx := r.Context()
		log := slogx.FromContext(ctx).With(slog.String("op", op))

		var reqBody request.SumRequest
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

		filter, err := reqBody.ToFilter()
		if err != nil {
			log.ErrorContext(ctx, "invalid filter", slogx.Err(err))
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		total, err := subscription.Sum(ctx, filter)
		if err != nil {
			log.ErrorContext(ctx, "failed to calculate sum", slogx.Err(err))
			render.JSON(w, r, response.Internal("internal error"))
			return
		}

		render.JSON(
			w, r,
			response.Response{
				Status: response.StatusOK,
				Data:   response.SumResponse{Total: total},
			},
		)
	}
}

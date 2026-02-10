package sumv1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/httpserver/request"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
	"github.com/salivare/subscriptions-service/internal/services/subscription"
)

// Subscription service interface
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
func New(s Subscription) http.HandlerFunc {
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

		if !request.ValidateStruct(w, r, &reqBody) {
			return
		}

		if reqBody.StartDateFrom == nil && reqBody.EndDateFrom == nil {
			render.JSON(
				w, r, response.Response{
					Status: response.StatusError,
					Error:  "either start_date_from or end_date_from must be provided",
					Code:   http.StatusBadRequest,
				},
			)
			return
		}

		filter, err := reqBody.ToFilter()
		if err != nil {
			log.ErrorContext(ctx, "invalid filter", slogx.Err(err))
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		total, err := s.Sum(ctx, filter)
		if err != nil {
			if errors.Is(err, subscription.ErrStartDateInFuture) ||
				errors.Is(err, subscription.ErrEndDateInFuture) {

				log.WarnContext(ctx, "invalid date range", slog.Any("error", err))
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			log.ErrorContext(ctx, "failed to calculate sum", slog.Any("error", err))
			render.JSON(w, r, response.Error("internal error"))
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

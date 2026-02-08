package savev1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
	"github.com/salivare/subscriptions-service/internal/lib/api/response"
)

type Request struct {
	ServiceName string `json:"service_name" validate:"required"`
	Price       int64  `json:"price" validate:"required"`
	UserID      string `json:"user_id" validate:"required"`
	StartDate   string `json:"start_date" validate:"required"`
}

type Response struct {
	response.Response
	Error string `json:"error,omitempty"`
}

type Subscription struct {
	ServiceName string
	Price       int64
	UserID      string
	StartDate   string
}

type SubscriptionSaver interface {
	SaveSubscription(subscription Subscription) (int64, error)
}

// New creates a handler for creating a subscription.
//
// @Summary Create subscription
// @Description Creates a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body Request true "Subscription data"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /api/v1/subscription [post]
func New(subscriptionSaver SubscriptionSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.subscriptions.save.New"
		log := slogx.FromContext(r.Context()).With(slog.String("op", op))

		var req Request
		if err := render.Bind(r, &req); err != nil {
			log.Error("invalid json", slogx.Err(err))
			render.JSON(w, r, response.Error("invalid json"))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			var validateError validator.ValidationErrors
			errors.As(err, &validateError)

			log.Error("invalid request", slogx.Err(err))

			render.JSON(w, r, response.ValidationError(validateError))
			return
		}

		// TODO: Добавить реализацию сохранения
		//_, err := subscriptionSaver.SaveSubscription(...)
		//if err != nil {
		//	log.Error("failed to save", slogx.Err(err))
		//	render.JSON(w, r, response.Error("failed to add subscription"))
		//	return
		//}

		render.JSON(
			w, r, Response{
				Response: response.OK(),
			},
		)
	}
}

package response

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	"github.com/salivare/subscriptions-service/internal/format"
)

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

const ErrNotFound = "subscription not found"

type Response struct {
	Status string      `json:"status"`
	Error  interface{} `json:"error,omitempty"`
	Data   interface{} `json:"data,omitempty"`
	Code   int         `json:"-"`
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       *int64    `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

type SumResponse struct {
	Total int64 `json:"total"`
}

func (r Response) StatusCode() int {
	if r.Code == 0 {
		return 200
	}

	return r.Code
}

func OK() Response {
	return Response{
		Status: StatusOK,
		Code:   http.StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
		Code:   http.StatusBadRequest,
	}
}

func Conflict(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
		Code:   http.StatusConflict,
	}
}

func Internal(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
		Code:   http.StatusInternalServerError,
	}
}

func ToSubscriptionResponse(m models.Subscription) SubscriptionResponse {
	var endDate *string
	if m.EndDate != nil {
		s := m.EndDate.Format(format.MonthYear)
		endDate = &s
	}

	return SubscriptionResponse{
		ID:          m.ID,
		ServiceName: m.ServiceName,
		Price:       m.Price,
		UserID:      m.UserID,
		StartDate:   m.StartDate.Format(format.MonthYear),
		EndDate:     endDate,
		CreatedAt:   m.CreatedAt.Format(time.DateTime),
		UpdatedAt:   m.UpdatedAt.Format(time.DateTime),
	}
}

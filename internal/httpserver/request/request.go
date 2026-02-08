package request

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/salivare/subscriptions-service/internal/domain/models"
)

type Request struct {
	ServiceName string `json:"service_name" validate:"required"`
	Price       int64  `json:"price" validate:"required,min=0"`
	UserID      string `json:"user_id" validate:"required,uuid"`
	StartDate   string `json:"start_date" validate:"required"`
	EndDate     string `json:"end_date"`
}

func (r Request) ToModel() (models.Subscription, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return models.Subscription{}, fmt.Errorf("invalid user_id: %w", err)
	}

	start, err := parseMonthYear(r.StartDate)
	if err != nil {
		return models.Subscription{}, fmt.Errorf("invalid start_date: %w", err)
	}

	var end *time.Time
	if r.EndDate != "" {
		t, err := parseMonthYear(r.EndDate)
		if err != nil {
			return models.Subscription{}, fmt.Errorf("invalid end_date: %w", err)
		}
		end = &t
	}

	return models.Subscription{
		ServiceName: r.ServiceName,
		Price:       r.Price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

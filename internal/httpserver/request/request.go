package request

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/salivare/subscriptions-service/internal/domain/models"
)

type CreateRequest struct {
	ServiceName string `json:"service_name" validate:"required"`
	Price       *int64 `json:"price" validate:"required,min=0"`
	UserID      string `json:"user_id" validate:"required,uuid"`
	StartDate   string `json:"start_date" validate:"required"`
	EndDate     string `json:"end_date"`
}

func (r CreateRequest) ToModel() (models.Subscription, error) {
	return convert(r.ServiceName, r.Price, r.UserID, r.StartDate, r.EndDate)
}

type UpdateRequest struct {
	ServiceName *string `json:"service_name"`
	Price       *int64  `json:"price" validate:"min=0"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

func (r UpdateRequest) ApplyTo(sub *models.Subscription) error {
	if r.ServiceName != nil {
		sub.ServiceName = *r.ServiceName
	}

	if r.Price != nil {
		sub.Price = r.Price
	}

	if r.StartDate != nil {
		t, err := parseMonthYear(*r.StartDate)
		if err != nil {
			return fmt.Errorf("invalid start_date: %w", err)
		}
		sub.StartDate = t
	}

	if r.EndDate != nil {
		if *r.EndDate == "" {
			sub.EndDate = nil
		} else {
			t, err := parseMonthYear(*r.EndDate)
			if err != nil {
				return fmt.Errorf("invalid end_date: %w", err)
			}
			sub.EndDate = &t
		}
	}

	return nil
}

func convert(service string, price *int64, userID string, start string, end string) (models.Subscription, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.Subscription{}, fmt.Errorf("invalid user_id: %w", err)
	}

	startDate, err := parseMonthYear(start)
	if err != nil {
		return models.Subscription{}, fmt.Errorf("invalid start_date: %w", err)
	}

	var endDate *time.Time
	if end != "" {
		t, err := parseMonthYear(end)
		if err != nil {
			return models.Subscription{}, fmt.Errorf("invalid end_date: %w", err)
		}
		endDate = &t
	}

	return models.Subscription{
		ServiceName: service,
		Price:       price,
		UserID:      uid,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

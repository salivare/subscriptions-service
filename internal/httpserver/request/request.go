package request

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	"github.com/salivare/subscriptions-service/internal/format"
)

type CreateRequest struct {
	ServiceName string `json:"service_name" validate:"required"`
	Price       *int64 `json:"price" validate:"required,min=0"`
	UserID      string `json:"user_id" validate:"required,uuid"`
	StartDate   string `json:"start_date" validate:"required"`
	EndDate     string `json:"end_date"`
}

type UpdateRequest struct {
	ServiceName *string `json:"service_name" validate:"omitempty,min=1"`
	Price       *int64  `json:"price" validate:"omitempty,min=1"`
	UserID      *string `json:"user_id" validate:"omitempty,uuid4"`
	StartDate   *string `json:"start_date" validate:"omitempty,datetime=01-2006"`
	EndDate     *string `json:"end_date" validate:"omitempty,datetime=02-2006"`
}

type SumRequest struct {
	UserID      *string `json:"user_id" validate:"omitempty,uuid4"`
	ServiceName *string `json:"service_name" validate:"omitempty,min=1"`

	StartDateFrom *string `json:"start_date_from" validate:"omitempty,datetime=01-2006"`
	StartDateTo   *string `json:"start_date_to" validate:"omitempty,datetime=01-2006"`

	EndDateFrom *string `json:"end_date_from" validate:"omitempty,datetime=01-2006"`
	EndDateTo   *string `json:"end_date_to" validate:"omitempty,datetime=01-2006"`
}

func (r CreateRequest) ToModel() (models.Subscription, error) {
	return convert(r.ServiceName, r.Price, r.UserID, r.StartDate, r.EndDate)
}

func (r SumRequest) ToFilter() (models.SumFilter, error) {
	parse := func(s *string) (*time.Time, error) {
		if s == nil {
			return nil, nil
		}
		t, err := time.Parse(format.MonthYear, *s)
		if err != nil {
			return nil, err
		}

		tt := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		return &tt, nil
	}

	startFrom, err := parse(r.StartDateFrom)
	if err != nil {
		return models.SumFilter{}, fmt.Errorf("invalid start_date_from: %w", err)
	}
	startTo, err := parse(r.StartDateTo)
	if err != nil {
		return models.SumFilter{}, fmt.Errorf("invalid start_date_to: %w", err)
	}
	endFrom, err := parse(r.EndDateFrom)
	if err != nil {
		return models.SumFilter{}, fmt.Errorf("invalid end_date_from: %w", err)
	}
	endTo, err := parse(r.EndDateTo)
	if err != nil {
		return models.SumFilter{}, fmt.Errorf("invalid end_date_to: %w", err)
	}

	return models.SumFilter{
		UserID:        r.UserID,
		ServiceName:   r.ServiceName,
		StartDateFrom: startFrom,
		StartDateTo:   startTo,
		EndDateFrom:   endFrom,
		EndDateTo:     endTo,
	}, nil
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
	t, err := time.Parse(format.MonthYear, s)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

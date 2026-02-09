package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       *int64
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SumFilter struct {
	UserID      *string
	ServiceName *string

	StartDateFrom *time.Time
	StartDateTo   *time.Time

	EndDateFrom *time.Time
	EndDateTo   *time.Time
}

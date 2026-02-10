package subscription

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"
	"github.com/salivare/subscriptions-service/internal/httpserver/request"
	"github.com/salivare/subscriptions-service/internal/storage"
)

var (
	ErrAlreadyExists     = errors.New("subscription already exists")
	ErrNotFound          = errors.New("subscription not found")
	ErrStartDateInFuture = errors.New("start_date_from cannot be in the future when start_date_to is omitted")
	ErrEndDateInFuture   = errors.New("end_date_from cannot be in the future when end_date_to is omitted")
)

// Saver Save Signature interface
type Saver interface {
	SaveSubscription(ctx context.Context, subscription models.Subscription) (uuid.UUID, time.Time, error)
}

// Updater Update Signature interface
type Updater interface {
	UpdateSubscription(ctx context.Context, subscription models.Subscription) (models.Subscription, error)
}

// Deleter Delete Signature interface
type Deleter interface {
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
}

// Getter Get Signature interface
type Getter interface {
	SubscriptionByID(ctx context.Context, id uuid.UUID) (models.Subscription, error)
}

// Summer Sum Signature interface
type Summer interface {
	SumSubscriptions(ctx context.Context, filter models.SumFilter) (int64, error)
}

type Service struct {
	subSaver   Saver
	subUpdater Updater
	subDeleter Deleter
	subGetter  Getter
	subSummer  Summer
}

// New Service constructor.
func New(
	subSaver Saver,
	subUpdater Updater,
	subDeleter Deleter,
	subGetter Getter,
	subSummer Summer,
) *Service {
	return &Service{
		subSaver:   subSaver,
		subUpdater: subUpdater,
		subDeleter: subDeleter,
		subGetter:  subGetter,
		subSummer:  subSummer,
	}
}

// Save implementation of the Subscription interface.
func (s *Service) Save(ctx context.Context, sub models.Subscription) (uuid.UUID, time.Time, error) {
	const op = "services.subscriptions.Create"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	id, createAt, err := s.subSaver.SaveSubscription(ctx, sub)
	if err != nil {
		if errors.Is(err, storage.ErrSubscriptionExists) {
			log.WarnContext(ctx, "subscription already exists", slogx.Err(err))
			return uuid.Nil, time.Time{}, ErrAlreadyExists
		}

		log.ErrorContext(ctx, "error creating subscription", slogx.Err(err))
		return uuid.Nil, time.Time{}, fmt.Errorf("create subscription: %w", err)
	}

	return id, createAt, nil
}

// Delete implementation of the Subscription interface.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "services.subscriptions.Delete"
	log := slogx.FromContext(ctx).With(slog.String("op", op), slog.String("id", id.String()))

	err := s.subDeleter.DeleteSubscription(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.WarnContext(ctx, "subscription not found", slogx.Err(err))
			return ErrNotFound
		}

		log.ErrorContext(ctx, "failed to delete subscription", slogx.Err(err))
		return err
	}

	return nil
}

// Update implementation of the Subscription interface.
func (s *Service) Update(ctx context.Context, id uuid.UUID, patch request.UpdateRequest) (models.Subscription, error) {
	const op = "services.subscriptions.Update"
	log := slogx.FromContext(ctx).With(
		slog.String("op", op),
		slog.String("id", id.String()),
	)

	current, err := s.subGetter.SubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.WarnContext(ctx, "subscription not found", slogx.Err(err))
			return models.Subscription{}, ErrNotFound
		}

		log.ErrorContext(ctx, "failed to get subscription", slogx.Err(err))
		return models.Subscription{}, fmt.Errorf("%s: get: %w", op, err)
	}

	if err := patch.ApplyTo(&current); err != nil {
		log.ErrorContext(ctx, "failed to apply patch", slogx.Err(err))
		return models.Subscription{}, fmt.Errorf("%s: apply: %w", op, err)
	}

	updated, err := s.subUpdater.UpdateSubscription(ctx, current)
	if err != nil {
		log.ErrorContext(ctx, "failed to update subscription", slogx.Err(err))
		return models.Subscription{}, fmt.Errorf("%s: update: %w", op, err)
	}

	log.InfoContext(ctx, "subscription updated")
	return updated, nil
}

// Get implementation of the Subscription interface.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (models.Subscription, error) {
	const op = "services.subscriptions.Get"
	log := slogx.FromContext(ctx).With(
		slog.String("op", op),
		slog.String("id", id.String()),
	)

	sub, err := s.subGetter.SubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.WarnContext(ctx, "subscription not found", slogx.Err(err))
			return models.Subscription{}, ErrNotFound
		}

		log.ErrorContext(ctx, "failed to get subscription", slogx.Err(err))
		return models.Subscription{}, fmt.Errorf("%s: get: %w", op, err)
	}

	return sub, nil
}

// Sum implementation of the Subscription interface.
func (s *Service) Sum(ctx context.Context, f models.SumFilter) (int64, error) {
	const op = "services.subscriptions.Get"
	log := slogx.FromContext(ctx).With(
		slog.String("op", op),
	)

	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	if f.StartDateFrom != nil && f.StartDateTo == nil {
		if f.StartDateFrom.After(currentMonth) {
			return 0, ErrStartDateInFuture
		}
		log.InfoContext(ctx, "auto-setting start_date_to to current month")
		f.StartDateTo = &currentMonth
	}

	if f.EndDateFrom != nil && f.EndDateTo == nil {
		if f.EndDateFrom.After(currentMonth) {
			return 0, ErrEndDateInFuture
		}
		log.InfoContext(ctx, "auto-setting end_date_to to current month")
		f.EndDateTo = &currentMonth
	}

	if f.StartDateFrom != nil && f.StartDateTo != nil {
		if f.StartDateTo.Before(*f.StartDateFrom) {
			log.WarnContext(
				ctx,
				"invalid start_date range",
				slog.Time("from", *f.StartDateFrom),
				slog.Time("to", *f.StartDateTo),
			)
			return 0, fmt.Errorf("start_date_to must be >= start_date_from")
		}
	}

	if f.EndDateFrom != nil && f.EndDateTo != nil {
		if f.EndDateTo.Before(*f.EndDateFrom) {
			log.WarnContext(
				ctx,
				"invalid end_date range",
				slog.Time("from", *f.EndDateFrom),
				slog.Time("to", *f.EndDateTo),
			)
			return 0, fmt.Errorf("end_date_to must be >= end_date_from")
		}
	}

	log.InfoContext(ctx, "calculating subscription sum")

	return s.subSummer.SumSubscriptions(ctx, f)
}

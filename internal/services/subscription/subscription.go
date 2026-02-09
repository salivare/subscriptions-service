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

var ErrAlreadyExists = errors.New("subscription already exists")
var ErrNotFound = errors.New("subscription not found")

type Saver interface {
	SaveSubscription(ctx context.Context, subscription models.Subscription) (uuid.UUID, time.Time, error)
}

type Updater interface {
	UpdateSubscription(ctx context.Context, subscription models.Subscription) (models.Subscription, error)
}

type Deleter interface {
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
}

type Getter interface {
	SubscriptionByID(ctx context.Context, id uuid.UUID) (models.Subscription, error)
}

type Service struct {
	subSaver   Saver
	subUpdater Updater
	subDeleter Deleter
	subGetter  Getter
}

func New(
	subSaver Saver,
	subUpdater Updater,
	subDeleter Deleter,
	subGetter Getter,
) *Service {
	return &Service{
		subSaver:   subSaver,
		subUpdater: subUpdater,
		subDeleter: subDeleter,
		subGetter:  subGetter,
	}
}

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

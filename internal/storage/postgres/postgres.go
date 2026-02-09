package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"

	"github.com/salivare/subscriptions-service/internal/config"
	"github.com/salivare/subscriptions-service/internal/storage"
)

const (
	PGErrUniqueViolation = "23505"
)

type Storage struct {
	db *sql.DB
}

func New(cfg config.PostgresConfig) (*Storage, error) {
	const op = "storage.postgres.New"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	var db *sql.DB

	err := storage.RetryBackoff(
		cfg.Retry, func() error {
			var err error
			db, err = sql.Open("pgx", dsn)
			if err != nil {
				return err
			}
			return db.Ping()
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveSubscription(ctx context.Context, sub models.Subscription) (uuid.UUID, error) {
	const op = "storage.postgres.SaveSubscription"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	query := `
        INSERT INTO subscriptions (
            service_name,
            price,
            user_id,
            start_date,
            end_date
        ) VALUES ($1, $2, $3, $4, $5)
        RETURNING id;
    `

	var id uuid.UUID

	err := s.db.QueryRowContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == PGErrUniqueViolation {
			log.WarnContext(ctx, "subscription already exists", slogx.Err(err))
			return uuid.Nil, fmt.Errorf("%s: %w", op, storage.ErrSubscriptionExists)
		}

		log.ErrorContext(ctx, "failed to save subscription", slogx.Err(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SubscriptionByID(ctx context.Context, id uuid.UUID) (models.Subscription, error) {
	const op = "storage.postgres.SubscriptionByID"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	query := `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions
        WHERE id = $1
    `

	var sub models.Subscription

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Subscription{}, storage.ErrNotFound
		}

		log.ErrorContext(ctx, "failed to get subscription", slogx.Err(err))
		return models.Subscription{}, fmt.Errorf("%s: %w", op, err)
	}

	return sub, nil
}

func (s *Storage) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	const op = "storage.postgres.DeleteSubscription"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	query := `DELETE FROM subscriptions WHERE id = $1`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		log.ErrorContext(ctx, "failed to delete subscription", slogx.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (s *Storage) UpdateSubscription(ctx context.Context, sub models.Subscription) error {
	const op = "storage.postgres.UpdateSubscription"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	query := `
        UPDATE subscriptions
        SET service_name = $1,
            price = $2,
            user_id = $3,
            start_date = $4,
            end_date = $5
        WHERE id = $6
    `

	res, err := s.db.ExecContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == PGErrUniqueViolation {
			return fmt.Errorf("%s: %w", op, storage.ErrSubscriptionExists)
		}

		log.ErrorContext(ctx, "failed to update subscription", slogx.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return storage.ErrNotFound
	}

	return nil
}

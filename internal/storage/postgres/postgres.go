package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salivare-io/slogx"
	"github.com/salivare/subscriptions-service/internal/domain/models"

	"github.com/salivare/subscriptions-service/internal/config"
	"github.com/salivare/subscriptions-service/internal/storage"
)

const (
	PGErrUniqueViolation = "23505"
)

type Storage struct {
	pool *pgxpool.Pool
}

// New Storage constructor.
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

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse DSN: %w", op, err)
	}

	config.MaxConns = int32(cfg.MaxConns)
	config.MinConns = int32(cfg.MinConns)
	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod

	var pool *pgxpool.Pool

	err = storage.RetryBackoff(
		cfg.Retry, func() error {
			pool, err = pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				return err
			}
			return pool.Ping(context.Background())
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

// SaveSubscription implementation of the Saver interface.
func (s *Storage) SaveSubscription(ctx context.Context, sub models.Subscription) (uuid.UUID, time.Time, error) {
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
        RETURNING id, created_at;
    `

	var (
		id        uuid.UUID
		createdAt time.Time
	)

	err := s.pool.QueryRow(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&id, &createdAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == PGErrUniqueViolation {
			log.WarnContext(ctx, "subscription already exists", slogx.Err(err))
			return uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, storage.ErrSubscriptionExists)
		}

		log.ErrorContext(ctx, "failed to save subscription", slogx.Err(err))
		return uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, createdAt, nil
}

// SubscriptionByID implementation of the Getter interface.
func (s *Storage) SubscriptionByID(ctx context.Context, id uuid.UUID) (models.Subscription, error) {
	const op = "storage.postgres.GetSubscription"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	query := `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions
        WHERE id = $1
    `

	var sub models.Subscription

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Subscription{}, storage.ErrNotFound
		}

		log.ErrorContext(ctx, "failed to get subscription", slogx.Err(err))
		return models.Subscription{}, fmt.Errorf("%s: %w", op, err)
	}

	return sub, nil
}

// DeleteSubscription implementation of the Deleter interface.
func (s *Storage) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	const op = "storage.postgres.DeleteSubscription"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	query := `DELETE FROM subscriptions WHERE id = $1`

	cmd, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		log.ErrorContext(ctx, "failed to delete subscription", slogx.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmd.RowsAffected() == 0 {
		log.ErrorContext(ctx, "subscription not found", slogx.Err(err))
		return storage.ErrNotFound
	}

	return nil
}

// UpdateSubscription implementation of the Updater interface.
func (s *Storage) UpdateSubscription(ctx context.Context, sub models.Subscription) (models.Subscription, error) {
	const op = "storage.postgres.UpdateSubscription"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	query := `
        UPDATE subscriptions
        SET
            service_name = $1,
            price        = $2,
            user_id      = $3,
            start_date   = $4,
            end_date     = $5
        WHERE id = $6
        RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at;
    `

	var updated models.Subscription

	err := s.pool.QueryRow(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	).Scan(
		&updated.ID,
		&updated.ServiceName,
		&updated.Price,
		&updated.UserID,
		&updated.StartDate,
		&updated.EndDate,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == PGErrUniqueViolation {
			log.WarnContext(ctx, "subscription already exists", slogx.Err(err))
			return models.Subscription{}, fmt.Errorf("%s: %w", op, storage.ErrSubscriptionExists)
		}

		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "subscription does not exist", slogx.Err(err))
			return models.Subscription{}, storage.ErrNotFound
		}

		log.ErrorContext(ctx, "failed to update subscription", slogx.Err(err))
		return models.Subscription{}, fmt.Errorf("%s: %w", op, err)
	}

	return updated, nil
}

// SumSubscriptions implementation of the Summer interface.
func (s *Storage) SumSubscriptions(ctx context.Context, f models.SumFilter) (int64, error) {
	const op = "storage.postgres.SumSubscriptions"
	log := slogx.FromContext(ctx).With(slog.String("op", op))

	var (
		conditions []string
		args       []any
		argIndex   = 1
	)

	add := func(cond string, val any) {
		conditions = append(conditions, fmt.Sprintf(cond, argIndex))
		args = append(args, val)
		argIndex++
	}

	if f.UserID != nil {
		add("user_id = $%d", *f.UserID)
	}

	if f.ServiceName != nil {
		add("service_name = $%d", *f.ServiceName)
	}

	if f.StartDateFrom != nil {
		add("start_date >= $%d", *f.StartDateFrom)
	}

	if f.StartDateTo != nil {
		add("start_date <= $%d", *f.StartDateTo)
	}

	if f.EndDateFrom != nil {
		add("end_date >= $%d", *f.EndDateFrom)
	}

	if f.EndDateTo != nil {
		add("end_date <= $%d", *f.EndDateTo)
	}

	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
    `

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := s.pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		log.ErrorContext(ctx, "failed to sum subscriptions", slogx.Err(err))
		return 0, fmt.Errorf("failed to execute sum query: %w", err)
	}

	return total, nil
}

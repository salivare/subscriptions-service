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

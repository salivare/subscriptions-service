package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/salivare/subscriptions-service/internal/config"
	"github.com/salivare/subscriptions-service/internal/storage"
)

func ensureDatabase(adminDSN, dbName string) error {
	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = db.Exec("CREATE DATABASE " + dbName)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	cfg := config.MustLoad()
	pg := cfg.Postgres

	baseDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%d",
		pg.User,
		pg.Password,
		pg.Host,
		pg.Port,
	)

	adminDSN := baseDSN + "/postgres?sslmode=" + pg.SSLMode

	if err := ensureDatabase(adminDSN, pg.DBName); err != nil {
		panic(fmt.Errorf("failed to ensure database: %w", err))
	}

	finalDSN := fmt.Sprintf(
		"%s/%s?sslmode=%s&x-migrations-table=%s",
		baseDSN,
		pg.DBName,
		pg.SSLMode,
		pg.MigrationsTable,
	)

	m, err := migrate.New("file://"+pg.MigrationsPath, finalDSN)
	if err != nil {
		panic(err)
	}

	err = storage.RetryBackoff(
		pg.Retry, func() error {
			err := m.Up()
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return nil
			}
			return err
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("applied migrations")
}

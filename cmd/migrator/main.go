package main

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/salivare/subscriptions-service/internal/config"
	"github.com/salivare/subscriptions-service/internal/storage"
)

func main() {
	cfg := config.MustLoad()
	pg := cfg.Postgres

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=%s",
		pg.User,
		pg.Password,
		pg.Host,
		pg.Port,
		pg.DBName,
		pg.SSLMode,
		pg.MigrationsTable,
	)

	m, err := migrate.New("file://"+pg.MigrationsPath, dsn)
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

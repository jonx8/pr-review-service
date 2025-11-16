package database

import (
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func RunMigrations(db *sqlx.DB) error {
	const method = "database.RunMigrations"

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		slog.Error("failed to create migration driver",
			"method", method,
			"error", err,
		)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		slog.Error("failed to create migration instance",
			"method", method,
			"error", err,
		)
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to apply migrations",
			"method", method,
			"error", err,
		)
		return err
	}

	slog.Info("migrations applied successfully",
		"method", method,
	)

	return nil
}

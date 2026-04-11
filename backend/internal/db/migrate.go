package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxv5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/config"
	embeddedmigrations "github.com/vasantharatnam/taskflow-sabbavarapu/backend/migrations"
)

func RunMigrations(cfg *config.Config) error {
	connConfig, err := pgx.ParseConfig(buildDSN(cfg))
	if err != nil {
		return fmt.Errorf("failed to parse db config for migrations: %w", err)
	}

	sqlDB := stdlib.OpenDB(*connConfig)
	defer sqlDB.Close()

	driver, err := pgxv5.WithInstance(sqlDB, &pgxv5.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration database driver: %w", err)
	}

	sourceDriver, err := iofs.New(embeddedmigrations.Files, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, cfg.DBName, driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

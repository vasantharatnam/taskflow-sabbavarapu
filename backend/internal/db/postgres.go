package db

import (
	"context"
	"fmt"
	"time"

    "github.com/jackc/pgx/v5/pgxpool"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/config"
)

func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	 dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		 cfg.DBUser,
		 cfg.DBPassword,
		 cfg.DBHost,
		 cfg.DBPort,
		 cfg.DBName,
		 cfg.DBSSLMode,
	)
	 
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx , dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return pool, nil
}
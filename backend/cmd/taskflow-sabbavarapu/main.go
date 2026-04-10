package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/auth"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/config"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/db"
	authHand "github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/handlers"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/middleware"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/repository"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting server", "env", cfg.AppEnv, "port", cfg.AppPort)

	pool, err := db.NewPostgresPool(cfg)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("database connected successfully")

	userRepo := repository.NewUserRepository(pool)
	authHandler := authHand.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)

	mux.Handle("/me", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "unauthorized",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"user_id": user.UserID,
			"email":   user.Email,
		})
	})))


	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited cleanly")
}

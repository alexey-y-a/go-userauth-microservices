package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-y-a/go-userauth-microservices/libs/logger"
	"github.com/alexey-y-a/go-userauth-microservices/libs/metrics"
	authconfig "github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/config"
	httpHandlers "github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/http"
	pgstorage "github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.Init()
	log := logger.L().With().Str("service", "auth-service").Logger()

	cfg := authconfig.LoadConfig()

	pgStore, err := pgstorage.NewStore(pgstorage.Config{
		DSN: cfg.PostgresDSN,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to init postgres store")
		os.Exit(1)
	}
	defer func() {
		if err := pgStore.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close postgres store")
		}
	}()

	authHandler := httpHandlers.NewAuthHandler(pgStore)

	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Error().Err(err).Msg("failed to write health response")
			return
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	instrumentedHandler := metrics.InstrumentHandler("auth-service", mux)

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      instrumentedHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("starting auth-service")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("auth-service stopped with error")
		}
	}()

	sig := <-sigChan
	log.Info().Str("signal", sig.String()).Msg("received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("auth-service graceful shutdown failed")
	} else {
		log.Info().Msg("auth-service stopped gracefully")
	}
}

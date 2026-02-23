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
	httpHandlers "github.com/alexey-y-a/go-userauth-microservices/services/user-service/internal/http"
	pgstorage "github.com/alexey-y-a/go-userauth-microservices/services/user-service/internal/storage/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.Init()

	log := logger.L().With().Str("service", "user-service").Logger()

	userDSN := os.Getenv("USER_SERVICE_DSN")
	if userDSN == "" {
		userDSN = "postgres://userauth:password@localhost:5432/userauth?sslmode=disable"
	}

	userStore, err := pgstorage.NewStore(pgstorage.Config{
		DSN: userDSN,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to init user-service postgres store")
		os.Exit(1)
	}
	defer func() {
		err := userStore.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close user-service postgres store")
		}
	}()

	mux := http.NewServeMux()

	userHandler := httpHandlers.NewUserHandler(userStore)
	userHandler.RegisterRoutes(mux)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Error().Err(err).Msg("failed to write health response")
			return
		}
	})

	addr := ":8081"

	mux.Handle("/metrics", promhttp.Handler())

	handlerWithRequestID := httpHandlers.RequestIDMiddleware(log, mux)

	instrumentedHandler := metrics.InstrumentHandler("user-service", handlerWithRequestID)

	server := &http.Server{
		Addr:         addr,
		Handler:      instrumentedHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", addr).Msg("starting user-service")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("user-service stopped with error")
		}
	}()

	sig := <-sigChan

	log.Info().Str("signal", sig.String()).Msg("received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("user-service graceful shutdown failed")
	} else {
		log.Info().Msg("user-service stopped gracefully")
	}
}

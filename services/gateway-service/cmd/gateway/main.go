package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-y-a/go-userauth-microservices/libs/logger"
	httpHandlers "github.com/alexey-y-a/go-userauth-microservices/services/gateway-service/internal/http"
)

func main() {
    logger.Init()

    log := logger.L().With().Str("service", "geteway-service").Logger()

    mux := http.NewServeMux()

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(http.StatusOK)
      _, err := w.Write([]byte(`{"status":"ok"}`))
      if err != nil {
          log.Error().Err(err).Msg("failed to write health response")
          return
      }
    })

    gatewayHandler := httpHandlers.NewGatewayHandler()
    gatewayHandler.RegisterRoutes(mux)

    addr := ":8082"

    server := &http.Server {
        Addr: addr,
        Handler: mux,
    }

    sigChan := make(chan os.Signal, 1)

    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func () {
        log.Info().Str("addr", addr).Msg("starting gateway-service")
        err := server.ListenAndServe()
        if err != nil && err != http.ErrServerClosed {
             log.Error().Err(err).Msg("gatwey-service stopped with error")
        }
    }()

    sig := <- sigChan

    log.Info().Str("signal", sig.String()).Msg("received shutdown signal")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err := server.Shutdown(ctx)
    if err != nil {
        log.Error().Err(err).Msg("gateway-service graceful shutdown failed")
    } else {
        log.Info().Msg("gateway-service stopped gracefully")
    }
}

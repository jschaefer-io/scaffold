package scaffold

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func Boot(ctx context.Context, logger *slog.Logger) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	mux := http.NewServeMux()
	handler := applyRoutes(mux, logger)
	srv := &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", "8080"),
		Handler: handler,
	}

	// Register routes
	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error listening and serving", "error", err)
		}
	}()

	// Handle Shutdown
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, shutdownCancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("unable to shutdown server", "error", err)
		}
	}()
	wg.Wait()

	return nil
}

package scaffold

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type middleware func(h http.Handler) http.Handler

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (t *statusResponseWriter) WriteHeader(statusCode int) {
	t.statusCode = statusCode
	t.ResponseWriter.WriteHeader(statusCode)
}

func loggerMiddleware(logger *slog.Logger, excludePaths []string) middleware {
	// Create a lookup map for the paths to exclude from logging
	filterLookup := make(map[any]struct{})
	for _, srv := range excludePaths {
		filterLookup[srv] = struct{}{}
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for excluded paths
			if _, ok := filterLookup[r.URL.Path]; ok {
				h.ServeHTTP(w, r)
				return
			}
			start := time.Now()
			wrapped := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			h.ServeHTTP(wrapped, r)

			logTarget := logger.InfoContext
			if wrapped.statusCode >= 500 {
				logTarget = logger.WarnContext
			}
			logTarget(
				r.Context(),
				"request",
				"method", r.Method,
				"path", r.URL.Path,
				"code", wrapped.statusCode,
				"durationMs", time.Since(start).Milliseconds(),
			)
		})
	}
}

func recoverMiddleware(logger *slog.Logger) middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					logger.ErrorContext(r.Context(), "recovered from panic",
						"error", err,
						"request", r.URL.Path,
						"method", r.Method,
						"stack", string(debug.Stack()),
					)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}

func requestIdMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resolve request id or generate a new one
		requestId := r.Header.Get("X-Request-Id")
		if requestId == "" {
			requestId = uuid.New().String()
		}

		// Add the request id to the response as well
		w.Header().Set("X-Request-Id", requestId)

		// pass request id with the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "requestId", requestId)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

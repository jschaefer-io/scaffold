package scaffold

import (
	"log/slog"
	"net/http"
)

func applyRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
) http.Handler {

	mux.HandleFunc("GET /test/recover", func(writer http.ResponseWriter, request *http.Request) {
		panic("test recover")
	})

	// Main system status routing group
	statusRouter := http.NewServeMux()
	mux.Handle("/system-status/", http.StripPrefix("/system-status", statusRouter))
	statusRouter.Handle("GET /ready", handleSystemReady())

	// Catch all 404
	mux.Handle("/", handle404())

	// Apply Global Middlewares
	return requestIdMiddleware(
		loggerMiddleware(logger, []string{"/system-status/ready"})(
			recoverMiddleware(logger)(
				mux,
			),
		),
	)
}

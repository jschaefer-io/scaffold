package main

import (
	"context"
	"log/slog"
	"os"
)

func newSlogLogger(ctxValueKeys [][2]string) *slog.Logger {
	defaultHandler := slog.NewJSONHandler(os.Stdout, nil)
	handler := contextLogger{
		Handler: defaultHandler,
		keys:    ctxValueKeys,
	}
	return slog.New(handler)
}

type contextLogger struct {
	slog.Handler
	keys [][2]string
}

func (h contextLogger) Handle(ctx context.Context, r slog.Record) error {
	list := make([]slog.Attr, 0)
	for _, key := range h.keys {
		v := ctx.Value(key[0])
		if val, ok := v.(string); ok {
			list = append(list, slog.String(key[1], val))
		}
	}
	r.AddAttrs(list...)
	return h.Handler.Handle(ctx, r)
}

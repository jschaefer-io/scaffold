package main

import (
	"context"
	"os"

	"github.com/jschaefer-io/scaffold"
)

func main() {
	logger := newSlogLogger([][2]string{
		{"requestId", "rid"},
	})
	ctx := context.Background()
	if err := scaffold.Boot(ctx, logger); err != nil {
		logger.Error("unable to boot application", "error", err)
		os.Exit(1)
	}
}

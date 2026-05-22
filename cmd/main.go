package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"modular_monolith/internal/bootstrap"
	"modular_monolith/internal/platform/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	if err := bootstrap.Run(ctx, cfg); err != nil {
		slog.Error("run server", "error", err)
		os.Exit(1)
	}
}

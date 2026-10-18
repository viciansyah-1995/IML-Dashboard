package main

import (
	"context"
	"iml/internal/app"
	"iml/internal/config"
	"iml/internal/database"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	c, e := config.Load()
	if e != nil {
		slog.Error("configuration", "error", e)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	p, e := database.Open(ctx, c.DatabaseURL)
	cancel()
	if e != nil {
		slog.Error("startup", "error", e)
		os.Exit(1)
	}
	defer p.Close()
	var migrated bool
	p.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name='report_versions')").Scan(&migrated)
	if !migrated {
		slog.Error("migrations required: make migrate")
		os.Exit(1)
	}
	server := &http.Server{Addr: c.Addr, Handler: app.New(p, c).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 60 * time.Second, WriteTimeout: 120 * time.Second, IdleTimeout: 60 * time.Second}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	slog.Info("server ready", "address", c.Addr, "mode", c.Mode, "demo", c.Demo)
	if e = server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		slog.Error("server stopped", "error", e)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"queuelet/internal/api"
	"queuelet/internal/store"
	"syscall"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	s, e := store.Open(os.Getenv("DATABASE_URL"))
	if s == nil {
		log.Fatal(e)
	}
	defer s.DB.Close()
	// A database that is not reachable yet is an outage to wait out, not a fatal error.
	for e != nil {
		log.Printf("database unavailable at startup, retrying: %v", e)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
		e = s.DB.PingContext(ctx)
	}
	files, e := filepath.Glob("migrations/*.sql")
	if e != nil {
		log.Fatal(e)
	}
	for _, file := range files {
		data, e := os.ReadFile(file)
		if e != nil {
			log.Fatal(e)
		}
		if _, e = s.DB.Exec(string(data)); e != nil {
			log.Fatal(e)
		}
	}
	app := api.New(s)
	go app.Loop(ctx)
	server := &http.Server{Addr: ":8200", Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() { <-ctx.Done(); server.Shutdown(context.Background()) }()
	log.Printf("queuelet listening on :8200")
	if e = server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		log.Fatal(e)
	}
}

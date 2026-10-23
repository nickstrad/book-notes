package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"queuelet/internal/client"
	"queuelet/internal/worker"
	"strings"
	"syscall"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	name := os.Getenv("WORKER_NAME")
	q := os.Getenv("WORKER_QUEUE")
	if q == "" {
		q = "emails"
	}
	kinds := os.Getenv("WORKER_KINDS")
	if kinds == "" {
		kinds = "echo,sleep,fail,fail3,crash,http"
	}
	if e := worker.Run(ctx, client.Env(), name, q, strings.Split(kinds, ",")); e != nil {
		log.Fatal(e)
	}
}

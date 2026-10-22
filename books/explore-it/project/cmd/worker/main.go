package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"queuelet/internal/client"
	"queuelet/internal/worker"
	"strings"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	name := os.Getenv("WORKER_NAME")
	q := os.Getenv("WORKER_QUEUE")
	if q == "" {
		q = "emails"
	}
	kinds := os.Getenv("WORKER_KINDS")
	if kinds == "" {
		kinds = "echo,sleep,fail,crash,http"
	}
	if e := worker.Run(ctx, client.Env(), name, q, strings.Split(kinds, ",")); e != nil {
		log.Fatal(e)
	}
}

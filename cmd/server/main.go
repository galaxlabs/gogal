package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gogal/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.RunServer(ctx); err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}

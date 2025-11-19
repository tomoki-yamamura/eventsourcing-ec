package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tomoki-yamamura/eventsourcing-ec/container"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
)

func main() {
	fmt.Println("Starting Event Sourcing E-Commerce Application")

	// Load config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DI Container setup
	cont := container.NewContainer()
	if err := cont.Inject(ctx, cfg); err != nil {
		log.Fatalf("Failed to inject dependencies: %v", err)
	}

	// Start background workers
	go func() {
		if err := cont.OutboxPublisher.Start(ctx); err != nil {
			log.Printf("Outbox publisher stopped: %v", err)
		}
	}()
	log.Println("Background workers started successfully")

	// Handler layer setup (CQRS) - Cart only for now
	// TODO: Add cart handlers when implemented

	// Router setup - minimal for now
	mux := http.NewServeMux()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Server starting on port %s\n", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-c
	fmt.Println("\nReceived shutdown signal, gracefully shutting down...")

	// Cancel context to stop workers
	cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	fmt.Println("Application stopped")
}

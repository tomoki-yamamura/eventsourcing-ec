package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tomoki-yamamura/eventsourcing-ec/container"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/register"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cont := container.NewContainer()
	if err := cont.Inject(ctx, cfg); err != nil {
		log.Fatalf("Failed to inject dependencies: %v", err)
	}

	go func() {
		if err := cont.OutboxPublisher.Start(ctx); err != nil {
			log.Printf("Outbox publisher stopped: %v", err)
		}
	}()

	go func() {
		if err := cont.ProjectorService.Start(ctx); err != nil {
			log.Printf("Projector service stopped: %v", err)
		}
	}()

	go func() {
		if err := cont.CartAbandonmentService.Start(ctx); err != nil {
			log.Printf("Cart abandonment service stopped: %v", err)
		}
	}()
	log.Println("Background workers started successfully")

	handlerRegister := register.NewHandlerRegister(cont)
	appRouter := handlerRegister.SetupRouter()
	mux := appRouter.SetupRoutes()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-c
	log.Println("\nReceived shutdown signal, gracefully shutting down...")
	cancel()

	if err := cont.CartAbandonmentService.Close(); err != nil {
		log.Printf("Cart abandonment service close error: %v", err)
	}

	if err := cont.ProjectorService.Close(); err != nil {
		log.Printf("Projector service close error: %v", err)
	}

	// Shutdown HTTP server
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Application stopped")
}

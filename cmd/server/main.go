package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rr/web/internal/auth"
	"rr/web/internal/config"
	"rr/web/internal/db"
	"rr/web/internal/handler"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config:", err)
	}

	// Initialize database
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize auth service
	authService := auth.NewService(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.CallbackURL,
		cfg.SessionSecret,
	)

	// Initialize handlers
	h := handler.New(database, authService, logger)

	// Create server
	srv := &http.Server{
		Addr:         ":3000",
		Handler:      h.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	// Create shutdown context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Println("Server stopped gracefully")
}

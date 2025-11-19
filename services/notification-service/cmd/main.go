package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshbarros/golang-airport-services/pkg/config"
	"github.com/joshbarros/golang-airport-services/pkg/database"
	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/infrastructure/notification/email"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/infrastructure/notification/push"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/infrastructure/notification/sms"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Initialize logger
	log, err := logger.New(cfg.Environment, "notification-service")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting Notification Service",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database
	db, err := database.NewPostgresDB(cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("Database connection established",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.Name),
	)

	// TODO: Initialize notification repository
	// For now, we'll use nil and focus on event consumption
	var notificationRepo repository.NotificationRepository = nil

	// Initialize notification providers
	providers := []repository.NotificationProvider{
		email.NewSMTPProvider(
			os.Getenv("SMTP_HOST"),
			os.Getenv("SMTP_PORT"),
			os.Getenv("SMTP_USERNAME"),
			os.Getenv("SMTP_PASSWORD"),
			os.Getenv("SMTP_FROM"),
			os.Getenv("SMTP_USE_TLS") == "true",
			log,
		),
		sms.NewConsoleProvider(log),
		push.NewConsoleProvider(log),
	}

	log.Info("Notification providers initialized",
		zap.Int("provider_count", len(providers)),
	)

	// Initialize use cases
	sendNotificationUC := usecase.NewSendNotificationUseCase(notificationRepo, providers)

	// Initialize RabbitMQ event consumer
	if cfg.RabbitMQ.URL == "" {
		log.Fatal("RABBITMQ_URL is required for notification service")
	}

	consumer, err := rabbitmq.NewFlightEventConsumer(
		cfg.RabbitMQ.URL,
		cfg.RabbitMQ.Exchange,
		sendNotificationUC,
		log,
	)
	if err != nil {
		log.Fatal("Failed to initialize event consumer", zap.Error(err))
	}
	defer consumer.Stop()

	// Start consuming events
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		log.Fatal("Failed to start event consumer", zap.Error(err))
	}

	log.Info("Event consumer started, listening for flight events...")

	// Simple HTTP server for health checks
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Info("Starting HTTP server",
			zap.String("address", srv.Addr),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down service...")

	// Cancel context to stop consumer
	cancel()

	// Graceful shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Service shutdown complete")
}

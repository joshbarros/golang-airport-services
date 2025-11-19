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
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/event"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/infrastructure/persistence/postgres"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/interfaces/http/handler"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/interfaces/http/router"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Initialize logger
	log, err := logger.New(cfg.Environment, "flight-service")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting Flight Service",
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

	// Initialize repository
	flightRepo := postgres.NewFlightRepository(db.Pool, log)

	// Initialize event publisher (optional - can be nil)
	var eventPublisher event.Publisher
	if cfg.RabbitMQ.URL != "" {
		var err error
		eventPublisher, err = rabbitmq.NewEventPublisher(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange, log)
		if err != nil {
			log.Warn("Failed to connect to RabbitMQ, events will not be published",
				zap.Error(err),
				zap.String("url", cfg.RabbitMQ.URL),
			)
		} else {
			defer eventPublisher.Close()
			log.Info("Event publisher initialized",
				zap.String("exchange", cfg.RabbitMQ.Exchange),
			)
		}
	} else {
		log.Info("RabbitMQ URL not configured, events will not be published")
	}

	// Initialize use cases (Dependency Injection)
	createFlightUC := usecase.NewCreateFlightUseCase(flightRepo, eventPublisher)
	getFlightUC := usecase.NewGetFlightUseCase(flightRepo)
	listFlightsUC := usecase.NewListFlightsUseCase(flightRepo)
	searchFlightsByNumberUC := usecase.NewSearchFlightsByNumberUseCase(flightRepo)
	updateFlightStatusUC := usecase.NewUpdateFlightStatusUseCase(flightRepo, eventPublisher)
	delayFlightUC := usecase.NewDelayFlightUseCase(flightRepo, eventPublisher)
	cancelFlightUC := usecase.NewCancelFlightUseCase(flightRepo, eventPublisher)

	// Initialize HTTP handler
	flightHandler := handler.NewFlightHandler(
		createFlightUC,
		getFlightUC,
		listFlightsUC,
		searchFlightsByNumberUC,
		updateFlightStatusUC,
		delayFlightUC,
		cancelFlightUC,
		log,
	)

	// Setup router
	r := router.SetupRouter(flightHandler, log, cfg.IsDevelopment())

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Starting HTTP server",
			zap.String("address", srv.Addr),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server shutdown complete")
}

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

	// Initialize use cases (Dependency Injection)
	createFlightUC := usecase.NewCreateFlightUseCase(flightRepo)
	getFlightUC := usecase.NewGetFlightUseCase(flightRepo)
	updateFlightStatusUC := usecase.NewUpdateFlightStatusUseCase(flightRepo)
	delayFlightUC := usecase.NewDelayFlightUseCase(flightRepo)
	cancelFlightUC := usecase.NewCancelFlightUseCase(flightRepo)

	// Initialize HTTP handler
	flightHandler := handler.NewFlightHandler(
		createFlightUC,
		getFlightUC,
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

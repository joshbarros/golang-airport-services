package router

import (
	"github.com/gin-gonic/gin"
	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/pkg/middleware"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/interfaces/http/handler"
)

// SetupRouter configures and returns the HTTP router
func SetupRouter(
	flightHandler *handler.FlightHandler,
	log *logger.Logger,
	isDevelopment bool,
) *gin.Engine {
	// Set Gin mode
	if !isDevelopment {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS())
	router.Use(middleware.SecurityHeaders())

	// Health check endpoints (no rate limiting)
	router.GET("/health", middleware.HealthCheck())
	router.GET("/ready", func(c *gin.Context) {
		// TODO: Add database connectivity check
		middleware.HealthCheck()(c)
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Apply rate limiting to API routes
		v1.Use(middleware.RateLimiter(100, 200)) // 100 requests/second, burst 200

		// Flight routes
		flights := v1.Group("/flights")
		{
			flights.GET("", flightHandler.ListFlights)
			flights.POST("", flightHandler.CreateFlight)
			flights.GET("/:id", flightHandler.GetFlight)
			flights.PATCH("/:id/status", flightHandler.UpdateFlightStatus)
			flights.POST("/:id/delay", flightHandler.DelayFlight)
			flights.POST("/:id/cancel", flightHandler.CancelFlight)
		}
	}

	return router
}

package repository

import (
	"context"
	"time"

	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/entity"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

// FlightRepository defines the contract for flight persistence
// This is part of the domain layer (Dependency Inversion Principle)
// The implementation will be in the infrastructure/adapter layer
type FlightRepository interface {
	// Save persists a flight (create or update)
	Save(ctx context.Context, flight *entity.Flight) error

	// FindByID retrieves a flight by its ID
	FindByID(ctx context.Context, id string) (*entity.Flight, error)

	// FindByFlightNumber retrieves a flight by its flight number and date
	FindByFlightNumber(ctx context.Context, flightNumber valueobject.FlightNumber, date time.Time) (*entity.Flight, error)

	// FindAll retrieves all flights with pagination
	FindAll(ctx context.Context, limit, offset int) ([]*entity.Flight, error)

	// FindByStatus retrieves flights by status with pagination
	FindByStatus(ctx context.Context, status valueobject.FlightStatus, limit, offset int) ([]*entity.Flight, error)

	// FindByOrigin retrieves flights by origin airport
	FindByOrigin(ctx context.Context, origin string, limit, offset int) ([]*entity.Flight, error)

	// FindByDestination retrieves flights by destination airport
	FindByDestination(ctx context.Context, destination string, limit, offset int) ([]*entity.Flight, error)

	// FindByOriginAndDestination retrieves flights by route
	FindByOriginAndDestination(ctx context.Context, origin, destination string, limit, offset int) ([]*entity.Flight, error)

	// FindByDateRange retrieves flights within a date range
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Flight, error)

	// FindActiveFlights retrieves all active (not arrived/cancelled) flights
	FindActiveFlights(ctx context.Context, limit, offset int) ([]*entity.Flight, error)

	// FindDelayedFlights retrieves all delayed flights
	FindDelayedFlights(ctx context.Context, limit, offset int) ([]*entity.Flight, error)

	// Delete removes a flight (soft delete recommended in production)
	Delete(ctx context.Context, id string) error

	// Count returns total number of flights
	Count(ctx context.Context) (int64, error)

	// CountByStatus returns count of flights by status
	CountByStatus(ctx context.Context, status valueobject.FlightStatus) (int64, error)
}

// FlightRepositoryError represents repository-specific errors
type FlightRepositoryError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
}

func (e *FlightRepositoryError) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e *FlightRepositoryError) Unwrap() error {
	return e.Err
}

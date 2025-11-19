package usecase

import (
	"context"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

// SearchFlightsByNumberInput represents the input for searching flights by number
type SearchFlightsByNumberInput struct {
	FlightNumber string
	Date         time.Time // Optional: search for specific date
}

// SearchFlightsByNumberUseCase handles searching flights by flight number
type SearchFlightsByNumberUseCase struct {
	flightRepo repository.FlightRepository
}

// NewSearchFlightsByNumberUseCase creates a new instance
func NewSearchFlightsByNumberUseCase(flightRepo repository.FlightRepository) *SearchFlightsByNumberUseCase {
	return &SearchFlightsByNumberUseCase{
		flightRepo: flightRepo,
	}
}

// Execute searches for flights by flight number
func (uc *SearchFlightsByNumberUseCase) Execute(ctx context.Context, input SearchFlightsByNumberInput) (*GetFlightOutput, error) {
	// Validate and create flight number value object
	flightNumber, err := valueobject.NewFlightNumber(input.FlightNumber)
	if err != nil {
		return nil, err
	}

	// Use the provided date or default to today
	searchDate := input.Date
	if searchDate.IsZero() {
		searchDate = time.Now()
	}

	// Normalize to start of day for consistent searching
	searchDate = time.Date(
		searchDate.Year(),
		searchDate.Month(),
		searchDate.Day(),
		0, 0, 0, 0,
		searchDate.Location(),
	)

	// Search in repository
	flight, err := uc.flightRepo.FindByFlightNumber(ctx, flightNumber, searchDate)
	if err != nil {
		return nil, errors.InternalServerError("failed to search flight").Wrap(err)
	}

	if flight == nil {
		return nil, errors.NotFound("flight not found for number: " + input.FlightNumber)
	}

	return toFlightOutput(flight), nil
}

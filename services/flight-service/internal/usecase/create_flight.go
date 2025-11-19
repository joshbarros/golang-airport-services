package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/entity"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/event"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

// CreateFlightInput represents the input for creating a flight
type CreateFlightInput struct {
	FlightNumber  string
	Origin        string
	Destination   string
	DepartureTime time.Time
	ArrivalTime   time.Time
	AircraftType  string
	Gate          string
	Terminal      string
}

// CreateFlightOutput represents the output after creating a flight
type CreateFlightOutput struct {
	ID            string
	FlightNumber  string
	Origin        string
	Destination   string
	DepartureTime time.Time
	ArrivalTime   time.Time
	Status        string
	AircraftType  string
	Gate          string
	Terminal      string
	CreatedAt     time.Time
}

// CreateFlightUseCase handles the creation of a new flight
type CreateFlightUseCase struct {
	flightRepo     repository.FlightRepository
	eventPublisher event.Publisher
}

// NewCreateFlightUseCase creates a new instance of CreateFlightUseCase
func NewCreateFlightUseCase(flightRepo repository.FlightRepository, eventPublisher event.Publisher) *CreateFlightUseCase {
	return &CreateFlightUseCase{
		flightRepo:     flightRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute creates a new flight
func (uc *CreateFlightUseCase) Execute(ctx context.Context, input CreateFlightInput) (*CreateFlightOutput, error) {
	// Validate and create flight number value object
	flightNumber, err := valueobject.NewFlightNumber(input.FlightNumber)
	if err != nil {
		return nil, err
	}

	// Check if flight already exists for the same date
	flightDate := time.Date(
		input.DepartureTime.Year(),
		input.DepartureTime.Month(),
		input.DepartureTime.Day(),
		0, 0, 0, 0,
		input.DepartureTime.Location(),
	)

	existingFlight, err := uc.flightRepo.FindByFlightNumber(ctx, flightNumber, flightDate)
	if err == nil && existingFlight != nil {
		return nil, errors.Conflict(
			fmt.Sprintf("flight %s already exists for date %s", input.FlightNumber, flightDate.Format("2006-01-02")),
		)
	}

	// Create new flight entity
	flight, err := entity.NewFlight(
		flightNumber,
		input.Origin,
		input.Destination,
		input.DepartureTime,
		input.ArrivalTime,
		input.AircraftType,
		input.Gate,
	)
	if err != nil {
		return nil, err
	}

	// Save flight to repository
	if err := uc.flightRepo.Save(ctx, flight); err != nil {
		return nil, errors.InternalServerError("failed to create flight").Wrap(err)
	}

	// Publish FlightCreatedEvent (optional - service can work without events)
	if uc.eventPublisher != nil {
		evt := event.FlightCreatedEvent{
			ID:            flight.ID(),
			FlightNumber:  flight.FlightNumber().String(),
			Origin:        flight.Origin(),
			Destination:   flight.Destination(),
			DepartureTime: flight.DepartureTime(),
			ArrivalTime:   flight.ArrivalTime(),
			AircraftType:  flight.AircraftType(),
			Gate:          flight.Gate(),
			CreatedAt:     flight.CreatedAt(),
		}
		// Don't fail the operation if event publishing fails
		_ = uc.eventPublisher.Publish(ctx, evt)
	}

	// Return output
	return &CreateFlightOutput{
		ID:            flight.ID(),
		FlightNumber:  flight.FlightNumber().String(),
		Origin:        flight.Origin(),
		Destination:   flight.Destination(),
		DepartureTime: flight.DepartureTime(),
		ArrivalTime:   flight.ArrivalTime(),
		Status:        flight.Status().String(),
		AircraftType:  flight.AircraftType(),
		Gate:          flight.Gate(),
		Terminal:      flight.Terminal(),
		CreatedAt:     flight.CreatedAt(),
	}, nil
}

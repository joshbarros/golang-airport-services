package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/event"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
)

// DelayFlightInput represents the input for delaying a flight
type DelayFlightInput struct {
	FlightID      string
	DelayDuration time.Duration
	Reason        string
}

// DelayFlightUseCase handles delaying a flight
type DelayFlightUseCase struct {
	flightRepo     repository.FlightRepository
	eventPublisher event.Publisher
}

// NewDelayFlightUseCase creates a new instance
func NewDelayFlightUseCase(flightRepo repository.FlightRepository, eventPublisher event.Publisher) *DelayFlightUseCase {
	return &DelayFlightUseCase{
		flightRepo:     flightRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute delays a flight
func (uc *DelayFlightUseCase) Execute(ctx context.Context, input DelayFlightInput) (*GetFlightOutput, error) {
	if strings.TrimSpace(input.FlightID) == "" {
		return nil, errors.BadRequest("flight ID is required")
	}

	if input.DelayDuration <= 0 {
		return nil, errors.BadRequest("delay duration must be positive")
	}

	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.BadRequest("delay reason is required")
	}

	// Retrieve flight
	flight, err := uc.flightRepo.FindByID(ctx, input.FlightID)
	if err != nil {
		return nil, err
	}

	if flight == nil {
		return nil, errors.NotFound("flight not found")
	}

	// Delay flight using domain logic
	if err := flight.Delay(input.DelayDuration, input.Reason); err != nil {
		return nil, err
	}

	// Save updated flight
	if err := uc.flightRepo.Save(ctx, flight); err != nil {
		return nil, errors.InternalServerError("failed to delay flight").Wrap(err)
	}

	// Publish FlightDelayedEvent (optional - service can work without events)
	if uc.eventPublisher != nil {
		evt := event.FlightDelayedEvent{
			FlightID:      flight.ID(),
			FlightNumber:  flight.FlightNumber().String(),
			Origin:        flight.Origin(),
			Destination:   flight.Destination(),
			DelayDuration: int(input.DelayDuration.Minutes()),
			Reason:        input.Reason,
			NewDeparture:  flight.DepartureTime(),
			NewArrival:    flight.ArrivalTime(),
			DelayedAt:     flight.UpdatedAt(),
		}
		_ = uc.eventPublisher.Publish(ctx, evt)
	}

	return toFlightOutput(flight), nil
}

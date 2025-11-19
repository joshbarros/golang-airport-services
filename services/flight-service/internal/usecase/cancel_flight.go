package usecase

import (
	"context"
	"strings"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/event"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
)

// CancelFlightInput represents the input for cancelling a flight
type CancelFlightInput struct {
	FlightID string
	Reason   string
}

// CancelFlightUseCase handles cancelling a flight
type CancelFlightUseCase struct {
	flightRepo     repository.FlightRepository
	eventPublisher event.Publisher
}

// NewCancelFlightUseCase creates a new instance
func NewCancelFlightUseCase(flightRepo repository.FlightRepository, eventPublisher event.Publisher) *CancelFlightUseCase {
	return &CancelFlightUseCase{
		flightRepo:     flightRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute cancels a flight
func (uc *CancelFlightUseCase) Execute(ctx context.Context, input CancelFlightInput) (*GetFlightOutput, error) {
	if strings.TrimSpace(input.FlightID) == "" {
		return nil, errors.BadRequest("flight ID is required")
	}

	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.BadRequest("cancellation reason is required")
	}

	// Retrieve flight
	flight, err := uc.flightRepo.FindByID(ctx, input.FlightID)
	if err != nil {
		return nil, err
	}

	if flight == nil {
		return nil, errors.NotFound("flight not found")
	}

	// Cancel flight using domain logic
	if err := flight.Cancel(input.Reason); err != nil {
		return nil, err
	}

	// Save updated flight
	if err := uc.flightRepo.Save(ctx, flight); err != nil {
		return nil, errors.InternalServerError("failed to cancel flight").Wrap(err)
	}

	// Publish FlightCancelledEvent (optional - service can work without events)
	if uc.eventPublisher != nil {
		evt := event.FlightCancelledEvent{
			FlightID:     flight.ID(),
			FlightNumber: flight.FlightNumber().String(),
			Origin:       flight.Origin(),
			Destination:  flight.Destination(),
			Reason:       input.Reason,
			CancelledAt:  flight.UpdatedAt(),
		}
		_ = uc.eventPublisher.Publish(ctx, evt)
	}

	return toFlightOutput(flight), nil
}

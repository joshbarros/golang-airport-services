package usecase

import (
	"context"
	"strings"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

// UpdateFlightStatusInput represents the input for updating flight status
type UpdateFlightStatusInput struct {
	FlightID string
	Status   string
}

// UpdateFlightStatusUseCase handles updating a flight's status
type UpdateFlightStatusUseCase struct {
	flightRepo repository.FlightRepository
}

// NewUpdateFlightStatusUseCase creates a new instance
func NewUpdateFlightStatusUseCase(flightRepo repository.FlightRepository) *UpdateFlightStatusUseCase {
	return &UpdateFlightStatusUseCase{
		flightRepo: flightRepo,
	}
}

// Execute updates a flight's status
func (uc *UpdateFlightStatusUseCase) Execute(ctx context.Context, input UpdateFlightStatusInput) (*GetFlightOutput, error) {
	if strings.TrimSpace(input.FlightID) == "" {
		return nil, errors.BadRequest("flight ID is required")
	}

	// Retrieve flight
	flight, err := uc.flightRepo.FindByID(ctx, input.FlightID)
	if err != nil {
		return nil, err
	}

	if flight == nil {
		return nil, errors.NotFound("flight not found")
	}

	// Create new status value object
	newStatus, err := valueobject.NewFlightStatus(input.Status)
	if err != nil {
		return nil, err
	}

	// Update status using domain logic
	if err := flight.UpdateStatus(newStatus); err != nil {
		return nil, err
	}

	// Save updated flight
	if err := uc.flightRepo.Save(ctx, flight); err != nil {
		return nil, errors.InternalServerError("failed to update flight status").Wrap(err)
	}

	// Return updated flight
	return toFlightOutput(flight), nil
}

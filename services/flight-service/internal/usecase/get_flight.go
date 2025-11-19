package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
)

// GetFlightOutput represents the output for getting a flight
type GetFlightOutput struct {
	ID                  string
	FlightNumber        string
	Origin              string
	Destination         string
	DepartureTime       time.Time
	ArrivalTime         time.Time
	ActualDepartureTime *time.Time
	ActualArrivalTime   *time.Time
	Status              string
	AircraftType        string
	Gate                string
	Terminal            string
	DelayReason         string
	CancellationReason  string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// GetFlightUseCase handles retrieving a flight by ID
type GetFlightUseCase struct {
	flightRepo repository.FlightRepository
}

// NewGetFlightUseCase creates a new instance of GetFlightUseCase
func NewGetFlightUseCase(flightRepo repository.FlightRepository) *GetFlightUseCase {
	return &GetFlightUseCase{
		flightRepo: flightRepo,
	}
}

// Execute retrieves a flight by ID
func (uc *GetFlightUseCase) Execute(ctx context.Context, id string) (*GetFlightOutput, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.BadRequest("flight ID is required")
	}

	flight, err := uc.flightRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if flight == nil {
		return nil, errors.NotFound("flight not found")
	}

	return toFlightOutput(flight), nil
}

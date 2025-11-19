package usecase

import (
	"context"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

// ListFlightsInput represents the input for listing flights
type ListFlightsInput struct {
	Status      string
	Origin      string
	Destination string
	StartDate   *time.Time
	EndDate     *time.Time
	ActiveOnly  bool
	DelayedOnly bool
	Limit       int
	Offset      int
}

// ListFlightsOutput represents the output for listing flights
type ListFlightsOutput struct {
	Flights    []*GetFlightOutput
	TotalCount int64
	Page       int
	PageSize   int
}

// ListFlightsUseCase handles listing and filtering flights
type ListFlightsUseCase struct {
	flightRepo repository.FlightRepository
}

// NewListFlightsUseCase creates a new instance
func NewListFlightsUseCase(flightRepo repository.FlightRepository) *ListFlightsUseCase {
	return &ListFlightsUseCase{
		flightRepo: flightRepo,
	}
}

// Execute lists flights based on filters
func (uc *ListFlightsUseCase) Execute(ctx context.Context, input ListFlightsInput) (*ListFlightsOutput, error) {
	// Set default pagination
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	var flights []*GetFlightOutput
	var totalCount int64

	// Priority 1: Date range filter
	if input.StartDate != nil && input.EndDate != nil {
		flightEntities, err := uc.flightRepo.FindByDateRange(ctx, *input.StartDate, *input.EndDate, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list flights by date range").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		totalCount, _ = uc.flightRepo.Count(ctx)
	} else if input.DelayedOnly {
		// Priority 2: Delayed flights filter
		flightEntities, err := uc.flightRepo.FindDelayedFlights(ctx, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list delayed flights").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		delayedStatus, _ := valueobject.NewFlightStatus("delayed")
		totalCount, _ = uc.flightRepo.CountByStatus(ctx, delayedStatus)
	} else if input.ActiveOnly {
		// Priority 3: Active flights filter
		flightEntities, err := uc.flightRepo.FindActiveFlights(ctx, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list active flights").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		totalCount, _ = uc.flightRepo.Count(ctx)
	} else if input.Status != "" {
		// Priority 4: Status filter
		status, err := valueobject.NewFlightStatus(input.Status)
		if err != nil {
			return nil, err
		}
		flightEntities, err := uc.flightRepo.FindByStatus(ctx, status, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list flights by status").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		totalCount, _ = uc.flightRepo.CountByStatus(ctx, status)
	} else if input.Origin != "" && input.Destination != "" {
		// Priority 5: Route filter (origin + destination)
		flightEntities, err := uc.flightRepo.FindByOriginAndDestination(ctx, input.Origin, input.Destination, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list flights by route").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		totalCount, _ = uc.flightRepo.Count(ctx)
	} else if input.Origin != "" {
		// Priority 6: Origin filter
		flightEntities, err := uc.flightRepo.FindByOrigin(ctx, input.Origin, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list flights by origin").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		totalCount, _ = uc.flightRepo.Count(ctx)
	} else if input.Destination != "" {
		// Priority 7: Destination filter
		flightEntities, err := uc.flightRepo.FindByDestination(ctx, input.Destination, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list flights by destination").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		totalCount, _ = uc.flightRepo.Count(ctx)
	} else {
		// Default: List all flights
		flightEntities, err := uc.flightRepo.FindAll(ctx, input.Limit, input.Offset)
		if err != nil {
			return nil, errors.InternalServerError("failed to list flights").Wrap(err)
		}
		for _, f := range flightEntities {
			flights = append(flights, toFlightOutput(f))
		}
		totalCount, _ = uc.flightRepo.Count(ctx)
	}

	page := (input.Offset / input.Limit) + 1

	return &ListFlightsOutput{
		Flights:    flights,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   input.Limit,
	}, nil
}

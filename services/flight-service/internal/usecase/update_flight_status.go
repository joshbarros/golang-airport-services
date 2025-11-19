package usecase

import (
	"context"
	"strings"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/event"
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
	flightRepo     repository.FlightRepository
	eventPublisher event.Publisher
}

// NewUpdateFlightStatusUseCase creates a new instance
func NewUpdateFlightStatusUseCase(flightRepo repository.FlightRepository, eventPublisher event.Publisher) *UpdateFlightStatusUseCase {
	return &UpdateFlightStatusUseCase{
		flightRepo:     flightRepo,
		eventPublisher: eventPublisher,
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

	// Publish FlightStatusChangedEvent (optional - service can work without events)
	if uc.eventPublisher != nil {
		evt := event.FlightStatusChangedEvent{
			FlightID:     flight.ID(),
			FlightNumber: flight.FlightNumber().String(),
			OldStatus:    flight.Status().String(), // Note: This is the new status, we'd need to track old status
			NewStatus:    newStatus.String(),
			ChangedAt:    flight.UpdatedAt(),
		}
		// Special events for specific status transitions
		switch newStatus.String() {
		case "boarding":
			boardingEvt := event.FlightBoardingStartedEvent{
				FlightID:     flight.ID(),
				FlightNumber: flight.FlightNumber().String(),
				Gate:         flight.Gate(),
				Terminal:     flight.Terminal(),
				StartedAt:    flight.UpdatedAt(),
			}
			_ = uc.eventPublisher.Publish(ctx, boardingEvt)
		case "departed":
			actualDep := flight.ActualDepartureTime()
			if actualDep.IsZero() {
				actualDep = flight.UpdatedAt()
			}
			departedEvt := event.FlightDepartedEvent{
				FlightID:           flight.ID(),
				FlightNumber:       flight.FlightNumber().String(),
				Origin:             flight.Origin(),
				Destination:        flight.Destination(),
				ActualDeparture:    actualDep,
				ScheduledDeparture: flight.DepartureTime(),
				DepartedAt:         flight.UpdatedAt(),
			}
			_ = uc.eventPublisher.Publish(ctx, departedEvt)
		case "arrived":
			actualArr := flight.ActualArrivalTime()
			if actualArr.IsZero() {
				actualArr = flight.UpdatedAt()
			}
			arrivedEvt := event.FlightArrivedEvent{
				FlightID:         flight.ID(),
				FlightNumber:     flight.FlightNumber().String(),
				Origin:           flight.Origin(),
				Destination:      flight.Destination(),
				ActualArrival:    actualArr,
				ScheduledArrival: flight.ArrivalTime(),
				ArrivedAt:        flight.UpdatedAt(),
			}
			_ = uc.eventPublisher.Publish(ctx, arrivedEvt)
		}
		_ = uc.eventPublisher.Publish(ctx, evt)
	}

	// Return updated flight
	return toFlightOutput(flight), nil
}

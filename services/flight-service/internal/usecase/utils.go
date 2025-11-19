package usecase

import (
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/entity"
)

// toFlightOutput converts a flight entity to output DTO
func toFlightOutput(flight *entity.Flight) *GetFlightOutput {
	output := &GetFlightOutput{
		ID:                 flight.ID(),
		FlightNumber:       flight.FlightNumber().String(),
		Origin:             flight.Origin(),
		Destination:        flight.Destination(),
		DepartureTime:      flight.DepartureTime(),
		ArrivalTime:        flight.ArrivalTime(),
		Status:             flight.Status().String(),
		AircraftType:       flight.AircraftType(),
		Gate:               flight.Gate(),
		Terminal:           flight.Terminal(),
		DelayReason:        flight.DelayReason(),
		CancellationReason: flight.CancellationReason(),
		CreatedAt:          flight.CreatedAt(),
		UpdatedAt:          flight.UpdatedAt(),
	}

	// Handle optional actual times
	if !flight.ActualDepartureTime().IsZero() {
		actualDep := flight.ActualDepartureTime()
		output.ActualDepartureTime = &actualDep
	}

	if !flight.ActualArrivalTime().IsZero() {
		actualArr := flight.ActualArrivalTime()
		output.ActualArrivalTime = &actualArr
	}

	return output
}

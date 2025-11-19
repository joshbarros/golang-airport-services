package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

// Flight represents a flight aggregate root following DDD principles
type Flight struct {
	id                    string
	flightNumber          valueobject.FlightNumber
	origin                string
	destination           string
	departureTime         time.Time
	arrivalTime           time.Time
	actualDepartureTime   time.Time
	actualArrivalTime     time.Time
	status                valueobject.FlightStatus
	aircraftType          string
	gate                  string
	terminal              string
	delayReason           string
	cancellationReason    string
	createdAt             time.Time
	updatedAt             time.Time
}

// NewFlight creates a new flight with validation
func NewFlight(
	flightNumber valueobject.FlightNumber,
	origin, destination string,
	departureTime, arrivalTime time.Time,
	aircraftType, gate string,
) (*Flight, error) {
	// Validate inputs
	if err := validateFlightData(origin, destination, departureTime, arrivalTime); err != nil {
		return nil, err
	}

	now := time.Now()
	flight := &Flight{
		id:            uuid.New().String(),
		flightNumber:  flightNumber,
		origin:        strings.ToUpper(strings.TrimSpace(origin)),
		destination:   strings.ToUpper(strings.TrimSpace(destination)),
		departureTime: departureTime,
		arrivalTime:   arrivalTime,
		status:        valueobject.Scheduled(),
		aircraftType:  aircraftType,
		gate:          gate,
		createdAt:     now,
		updatedAt:     now,
	}

	return flight, nil
}

// validateFlightData validates flight creation data
func validateFlightData(origin, destination string, departureTime, arrivalTime time.Time) error {
	if strings.TrimSpace(origin) == "" {
		return errors.BadRequest("origin airport code is required")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.BadRequest("destination airport code is required")
	}

	if strings.ToUpper(strings.TrimSpace(origin)) == strings.ToUpper(strings.TrimSpace(destination)) {
		return errors.BadRequest("origin and destination cannot be the same")
	}

	if arrivalTime.Before(departureTime) || arrivalTime.Equal(departureTime) {
		return errors.BadRequest("arrival time must be after departure time")
	}

	return nil
}

// UpdateStatus updates the flight status with validation
func (f *Flight) UpdateStatus(newStatus valueobject.FlightStatus) error {
	// Check if transition is valid
	if !f.status.CanTransitionTo(newStatus) {
		return errors.BadRequest(
			fmt.Sprintf("cannot transition from %s to %s", f.status.String(), newStatus.String()),
		)
	}

	f.status = newStatus
	f.updatedAt = time.Now()
	return nil
}

// Delay delays the flight with a reason
func (f *Flight) Delay(duration time.Duration, reason string) error {
	// Cannot delay if already departed
	if f.status.Equals(valueobject.Departed()) ||
		f.status.Equals(valueobject.InFlight()) ||
		f.status.Equals(valueobject.Landed()) ||
		f.status.Equals(valueobject.Arrived()) {
		return errors.BadRequest("cannot delay flight after departure")
	}

	// Cannot delay if cancelled
	if f.status.Equals(valueobject.Cancelled()) {
		return errors.BadRequest("cannot delay cancelled flight")
	}

	f.departureTime = f.departureTime.Add(duration)
	f.arrivalTime = f.arrivalTime.Add(duration)
	f.status = valueobject.Delayed()
	f.delayReason = reason
	f.updatedAt = time.Now()

	return nil
}

// Cancel cancels the flight with a reason
func (f *Flight) Cancel(reason string) error {
	// Cannot cancel if already departed or in flight
	if f.status.Equals(valueobject.Departed()) ||
		f.status.Equals(valueobject.InFlight()) ||
		f.status.Equals(valueobject.Landed()) ||
		f.status.Equals(valueobject.Arrived()) {
		return errors.BadRequest("cannot cancel flight after departure")
	}

	// Cannot cancel if already cancelled
	if f.status.Equals(valueobject.Cancelled()) {
		return errors.BadRequest("flight is already cancelled")
	}

	f.status = valueobject.Cancelled()
	f.cancellationReason = reason
	f.updatedAt = time.Now()

	return nil
}

// UpdateGate updates the gate assignment
func (f *Flight) UpdateGate(gate string) error {
	// Cannot change gate after boarding has started
	if f.status.Equals(valueobject.Boarding()) ||
		f.status.Equals(valueobject.Departed()) ||
		f.status.Equals(valueobject.InFlight()) ||
		f.status.Equals(valueobject.Landed()) ||
		f.status.Equals(valueobject.Arrived()) {
		return errors.BadRequest("cannot change gate after boarding has started")
	}

	f.gate = gate
	f.updatedAt = time.Now()
	return nil
}

// SetActualDepartureTime sets the actual departure time
func (f *Flight) SetActualDepartureTime(t time.Time) error {
	if f.status.Equals(valueobject.Scheduled()) {
		return errors.BadRequest("cannot set actual departure time before boarding")
	}

	f.actualDepartureTime = t
	f.updatedAt = time.Now()
	return nil
}

// SetActualArrivalTime sets the actual arrival time
func (f *Flight) SetActualArrivalTime(t time.Time) error {
	if !f.status.Equals(valueobject.Landed()) && !f.status.Equals(valueobject.Arrived()) {
		return errors.BadRequest("cannot set actual arrival time before landing")
	}

	f.actualArrivalTime = t
	f.updatedAt = time.Now()
	return nil
}

// IsDelayed returns true if the flight is delayed
func (f *Flight) IsDelayed() bool {
	return f.status.Equals(valueobject.Delayed())
}

// IsActive returns true if the flight is in an active state
func (f *Flight) IsActive() bool {
	return f.status.IsActive()
}

// Getters (following encapsulation principle)

func (f *Flight) ID() string {
	return f.id
}

func (f *Flight) FlightNumber() valueobject.FlightNumber {
	return f.flightNumber
}

func (f *Flight) Origin() string {
	return f.origin
}

func (f *Flight) Destination() string {
	return f.destination
}

func (f *Flight) DepartureTime() time.Time {
	return f.departureTime
}

func (f *Flight) ArrivalTime() time.Time {
	return f.arrivalTime
}

func (f *Flight) ActualDepartureTime() time.Time {
	return f.actualDepartureTime
}

func (f *Flight) ActualArrivalTime() time.Time {
	return f.actualArrivalTime
}

func (f *Flight) Status() valueobject.FlightStatus {
	return f.status
}

func (f *Flight) AircraftType() string {
	return f.aircraftType
}

func (f *Flight) Gate() string {
	return f.gate
}

func (f *Flight) Terminal() string {
	return f.terminal
}

func (f *Flight) DelayReason() string {
	return f.delayReason
}

func (f *Flight) CancellationReason() string {
	return f.cancellationReason
}

func (f *Flight) CreatedAt() time.Time {
	return f.createdAt
}

func (f *Flight) UpdatedAt() time.Time {
	return f.updatedAt
}

// Reconstruct recreates a flight from persistence (used by repository)
func Reconstruct(
	id string,
	flightNumber valueobject.FlightNumber,
	origin, destination string,
	departureTime, arrivalTime time.Time,
	actualDepartureTime, actualArrivalTime time.Time,
	status valueobject.FlightStatus,
	aircraftType, gate, terminal string,
	delayReason, cancellationReason string,
	createdAt, updatedAt time.Time,
) *Flight {
	return &Flight{
		id:                  id,
		flightNumber:        flightNumber,
		origin:              origin,
		destination:         destination,
		departureTime:       departureTime,
		arrivalTime:         arrivalTime,
		actualDepartureTime: actualDepartureTime,
		actualArrivalTime:   actualArrivalTime,
		status:              status,
		aircraftType:        aircraftType,
		gate:                gate,
		terminal:            terminal,
		delayReason:         delayReason,
		cancellationReason:  cancellationReason,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

package valueobject

import (
	"fmt"
	"strings"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
)

// FlightStatus represents the status of a flight
type FlightStatus struct {
	value string
}

// Valid flight statuses
const (
	FlightStatusScheduled = "scheduled"
	FlightStatusBoarding  = "boarding"
	FlightStatusDeparted  = "departed"
	FlightStatusInFlight  = "in_flight"
	FlightStatusLanded    = "landed"
	FlightStatusArrived   = "arrived"
	FlightStatusDelayed   = "delayed"
	FlightStatusCancelled = "cancelled"
)

var (
	validStatuses = map[string]bool{
		FlightStatusScheduled: true,
		FlightStatusBoarding:  true,
		FlightStatusDeparted:  true,
		FlightStatusInFlight:  true,
		FlightStatusLanded:    true,
		FlightStatusArrived:   true,
		FlightStatusDelayed:   true,
		FlightStatusCancelled: true,
	}

	// Status transition rules: from -> allowed destinations
	statusTransitions = map[string][]string{
		FlightStatusScheduled: {FlightStatusBoarding, FlightStatusDelayed, FlightStatusCancelled},
		FlightStatusDelayed:   {FlightStatusBoarding, FlightStatusCancelled},
		FlightStatusBoarding:  {FlightStatusDeparted, FlightStatusDelayed, FlightStatusCancelled},
		FlightStatusDeparted:  {FlightStatusInFlight},
		FlightStatusInFlight:  {FlightStatusLanded, FlightStatusDelayed},
		FlightStatusLanded:    {FlightStatusArrived},
		FlightStatusArrived:   {}, // Final state
		FlightStatusCancelled: {}, // Final state
	}
)

// NewFlightStatus creates a new FlightStatus value object
func NewFlightStatus(value string) (FlightStatus, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))

	if !validStatuses[normalized] {
		return FlightStatus{}, errors.BadRequest(
			fmt.Sprintf("invalid flight status: %s (allowed: scheduled, boarding, departed, in_flight, landed, arrived, delayed, cancelled)", value),
		)
	}

	return FlightStatus{value: normalized}, nil
}

// String returns the string representation of the status
func (fs FlightStatus) String() string {
	return fs.value
}

// Equals checks if two statuses are equal
func (fs FlightStatus) Equals(other FlightStatus) bool {
	return fs.value == other.value
}

// IsFinal returns true if this is a final status (no further transitions possible)
func (fs FlightStatus) IsFinal() bool {
	return fs.value == FlightStatusArrived || fs.value == FlightStatusCancelled
}

// IsActive returns true if the flight is in an active state (not final)
func (fs FlightStatus) IsActive() bool {
	return !fs.IsFinal()
}

// CanTransitionTo checks if this status can transition to another status
func (fs FlightStatus) CanTransitionTo(target FlightStatus) bool {
	allowedTransitions, exists := statusTransitions[fs.value]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == target.value {
			return true
		}
	}

	return false
}

// IsZero checks if the status is empty
func (fs FlightStatus) IsZero() bool {
	return fs.value == ""
}

// Scheduled returns a scheduled status
func Scheduled() FlightStatus {
	return FlightStatus{value: FlightStatusScheduled}
}

// Boarding returns a boarding status
func Boarding() FlightStatus {
	return FlightStatus{value: FlightStatusBoarding}
}

// Departed returns a departed status
func Departed() FlightStatus {
	return FlightStatus{value: FlightStatusDeparted}
}

// InFlight returns an in-flight status
func InFlight() FlightStatus {
	return FlightStatus{value: FlightStatusInFlight}
}

// Landed returns a landed status
func Landed() FlightStatus {
	return FlightStatus{value: FlightStatusLanded}
}

// Arrived returns an arrived status
func Arrived() FlightStatus {
	return FlightStatus{value: FlightStatusArrived}
}

// Delayed returns a delayed status
func Delayed() FlightStatus {
	return FlightStatus{value: FlightStatusDelayed}
}

// Cancelled returns a cancelled status
func Cancelled() FlightStatus {
	return FlightStatus{value: FlightStatusCancelled}
}

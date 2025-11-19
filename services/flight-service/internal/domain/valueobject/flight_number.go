package valueobject

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
)

// FlightNumber represents a flight number value object
// Format: 2-3 letter airline code + 3-4 digit number (e.g., AA123, DAL1234)
type FlightNumber struct {
	value   string
	airline string
	number  string
}

var (
	// flightNumberRegex validates flight number format
	flightNumberRegex = regexp.MustCompile(`^[A-Z]{2,3}\d{3,4}$`)
)

// NewFlightNumber creates a new FlightNumber value object
func NewFlightNumber(value string) (FlightNumber, error) {
	// Normalize to uppercase
	normalized := strings.ToUpper(strings.TrimSpace(value))

	// Validate format
	if !flightNumberRegex.MatchString(normalized) {
		return FlightNumber{}, errors.BadRequest(
			fmt.Sprintf("invalid flight number format: %s (expected format: 2-3 letters + 3-4 digits, e.g., AA123)", value),
		)
	}

	// Extract airline code and number
	var airline, number string
	for i, char := range normalized {
		if char >= '0' && char <= '9' {
			airline = normalized[:i]
			number = normalized[i:]
			break
		}
	}

	return FlightNumber{
		value:   normalized,
		airline: airline,
		number:  number,
	}, nil
}

// String returns the string representation of the flight number
func (fn FlightNumber) String() string {
	return fn.value
}

// Airline returns the airline code part of the flight number
func (fn FlightNumber) Airline() string {
	return fn.airline
}

// Number returns the numeric part of the flight number
func (fn FlightNumber) Number() string {
	return fn.number
}

// Equals checks if two flight numbers are equal
func (fn FlightNumber) Equals(other FlightNumber) bool {
	return fn.value == other.value
}

// IsZero checks if the flight number is empty
func (fn FlightNumber) IsZero() bool {
	return fn.value == ""
}

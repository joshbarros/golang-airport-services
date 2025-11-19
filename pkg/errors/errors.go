package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a custom application error with additional context
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Cause      error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements the errors.Unwrap interface for error wrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Metadata:   make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(metadata map[string]interface{}) *AppError {
	e.Metadata = metadata
	return e
}

// Wrap wraps an underlying error
func (e *AppError) Wrap(cause error) *AppError {
	e.Cause = cause
	return e
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// Common error constructors following REST/HTTP conventions

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) *AppError {
	return NewAppError("BAD_REQUEST", message, http.StatusBadRequest)
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *AppError {
	return NewAppError("UNAUTHORIZED", message, http.StatusUnauthorized)
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *AppError {
	return NewAppError("FORBIDDEN", message, http.StatusForbidden)
}

// NotFound creates a 404 Not Found error
func NotFound(message string) *AppError {
	return NewAppError("NOT_FOUND", message, http.StatusNotFound)
}

// Conflict creates a 409 Conflict error
func Conflict(message string) *AppError {
	return NewAppError("CONFLICT", message, http.StatusConflict)
}

// InternalServerError creates a 500 Internal Server Error
func InternalServerError(message string) *AppError {
	return NewAppError("INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError)
}

// Domain-specific error codes for airport services
const (
	// Flight errors
	ErrFlightNotFound       = "FLIGHT_NOT_FOUND"
	ErrFlightAlreadyExists  = "FLIGHT_ALREADY_EXISTS"
	ErrInvalidFlightNumber  = "INVALID_FLIGHT_NUMBER"
	ErrInvalidFlightStatus  = "INVALID_FLIGHT_STATUS"
	ErrFlightCannotBeDelayed = "FLIGHT_CANNOT_BE_DELAYED"
	ErrFlightCannotBeCancelled = "FLIGHT_CANNOT_BE_CANCELLED"

	// Booking errors
	ErrBookingNotFound      = "BOOKING_NOT_FOUND"
	ErrInvalidBookingData   = "INVALID_BOOKING_DATA"
	ErrBookingAlreadyExists = "BOOKING_ALREADY_EXISTS"
	ErrMaxPassengersExceeded = "MAX_PASSENGERS_EXCEEDED"

	// Passenger errors
	ErrPassengerNotFound    = "PASSENGER_NOT_FOUND"
	ErrInvalidPassengerData = "INVALID_PASSENGER_DATA"

	// Baggage errors
	ErrBaggageNotFound      = "BAGGAGE_NOT_FOUND"
	ErrInvalidBaggageWeight = "INVALID_BAGGAGE_WEIGHT"
	ErrBaggageTagDuplicate  = "BAGGAGE_TAG_DUPLICATE"

	// Gate errors
	ErrGateNotFound         = "GATE_NOT_FOUND"
	ErrGateAlreadyAssigned  = "GATE_ALREADY_ASSIGNED"
	ErrGateNotAvailable     = "GATE_NOT_AVAILABLE"

	// Check-in errors
	ErrCheckInNotFound      = "CHECKIN_NOT_FOUND"
	ErrCheckInAlreadyDone   = "CHECKIN_ALREADY_DONE"
	ErrCheckInWindowClosed  = "CHECKIN_WINDOW_CLOSED"

	// Authentication errors
	ErrInvalidCredentials   = "INVALID_CREDENTIALS"
	ErrTokenExpired         = "TOKEN_EXPIRED"
	ErrInvalidToken         = "INVALID_TOKEN"

	// Validation errors
	ErrValidationFailed     = "VALIDATION_FAILED"
	ErrRequiredField        = "REQUIRED_FIELD"
	ErrInvalidFormat        = "INVALID_FORMAT"
)

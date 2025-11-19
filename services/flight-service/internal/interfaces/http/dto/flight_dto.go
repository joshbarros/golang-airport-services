package dto

import "time"

// CreateFlightRequest represents the HTTP request body for creating a flight
type CreateFlightRequest struct {
	FlightNumber  string    `json:"flight_number" binding:"required"`
	Origin        string    `json:"origin" binding:"required,len=3"`
	Destination   string    `json:"destination" binding:"required,len=3"`
	DepartureTime time.Time `json:"departure_time" binding:"required"`
	ArrivalTime   time.Time `json:"arrival_time" binding:"required"`
	AircraftType  string    `json:"aircraft_type" binding:"required"`
	Gate          string    `json:"gate" binding:"required"`
	Terminal      string    `json:"terminal"`
}

// UpdateFlightStatusRequest represents the HTTP request body for updating flight status
type UpdateFlightStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// DelayFlightRequest represents the HTTP request body for delaying a flight
type DelayFlightRequest struct {
	DelayMinutes int    `json:"delay_minutes" binding:"required,min=1"`
	Reason       string `json:"reason" binding:"required"`
}

// CancelFlightRequest represents the HTTP request body for cancelling a flight
type CancelFlightRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// FlightResponse represents the HTTP response for a flight
type FlightResponse struct {
	ID                  string     `json:"id"`
	FlightNumber        string     `json:"flight_number"`
	Origin              string     `json:"origin"`
	Destination         string     `json:"destination"`
	DepartureTime       time.Time  `json:"departure_time"`
	ArrivalTime         time.Time  `json:"arrival_time"`
	ActualDepartureTime *time.Time `json:"actual_departure_time,omitempty"`
	ActualArrivalTime   *time.Time `json:"actual_arrival_time,omitempty"`
	Status              string     `json:"status"`
	AircraftType        string     `json:"aircraft_type"`
	Gate                string     `json:"gate"`
	Terminal            string     `json:"terminal,omitempty"`
	DelayReason         string     `json:"delay_reason,omitempty"`
	CancellationReason  string     `json:"cancellation_reason,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// FlightListResponse represents the HTTP response for a list of flights
type FlightListResponse struct {
	Flights    []FlightResponse `json:"flights"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

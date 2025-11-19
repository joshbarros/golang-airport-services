package event

import "time"

// DomainEvent represents a domain event that occurred
type DomainEvent interface {
	EventType() string
	EventID() string
	OccurredAt() time.Time
}

// FlightCreatedEvent represents a flight creation event
type FlightCreatedEvent struct {
	ID            string    `json:"id"`
	FlightNumber  string    `json:"flight_number"`
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	AircraftType  string    `json:"aircraft_type"`
	Gate          string    `json:"gate"`
	CreatedAt     time.Time `json:"created_at"`
}

func (e FlightCreatedEvent) EventType() string {
	return "flight.created"
}

func (e FlightCreatedEvent) EventID() string {
	return e.ID
}

func (e FlightCreatedEvent) OccurredAt() time.Time {
	return e.CreatedAt
}

// FlightStatusChangedEvent represents a flight status change event
type FlightStatusChangedEvent struct {
	FlightID     string    `json:"flight_id"`
	FlightNumber string    `json:"flight_number"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	ChangedAt    time.Time `json:"changed_at"`
}

func (e FlightStatusChangedEvent) EventType() string {
	return "flight.status.changed"
}

func (e FlightStatusChangedEvent) EventID() string {
	return e.FlightID
}

func (e FlightStatusChangedEvent) OccurredAt() time.Time {
	return e.ChangedAt
}

// FlightDelayedEvent represents a flight delay event
type FlightDelayedEvent struct {
	FlightID      string        `json:"flight_id"`
	FlightNumber  string        `json:"flight_number"`
	Origin        string        `json:"origin"`
	Destination   string        `json:"destination"`
	DelayDuration int           `json:"delay_minutes"` // in minutes
	Reason        string        `json:"reason"`
	NewDeparture  time.Time     `json:"new_departure_time"`
	NewArrival    time.Time     `json:"new_arrival_time"`
	DelayedAt     time.Time     `json:"delayed_at"`
}

func (e FlightDelayedEvent) EventType() string {
	return "flight.delayed"
}

func (e FlightDelayedEvent) EventID() string {
	return e.FlightID
}

func (e FlightDelayedEvent) OccurredAt() time.Time {
	return e.DelayedAt
}

// FlightCancelledEvent represents a flight cancellation event
type FlightCancelledEvent struct {
	FlightID     string    `json:"flight_id"`
	FlightNumber string    `json:"flight_number"`
	Origin       string    `json:"origin"`
	Destination  string    `json:"destination"`
	Reason       string    `json:"reason"`
	CancelledAt  time.Time `json:"cancelled_at"`
}

func (e FlightCancelledEvent) EventType() string {
	return "flight.cancelled"
}

func (e FlightCancelledEvent) EventID() string {
	return e.FlightID
}

func (e FlightCancelledEvent) OccurredAt() time.Time {
	return e.CancelledAt
}

// FlightBoardingStartedEvent represents boarding start event
type FlightBoardingStartedEvent struct {
	FlightID     string    `json:"flight_id"`
	FlightNumber string    `json:"flight_number"`
	Gate         string    `json:"gate"`
	Terminal     string    `json:"terminal"`
	StartedAt    time.Time `json:"started_at"`
}

func (e FlightBoardingStartedEvent) EventType() string {
	return "flight.boarding.started"
}

func (e FlightBoardingStartedEvent) EventID() string {
	return e.FlightID
}

func (e FlightBoardingStartedEvent) OccurredAt() time.Time {
	return e.StartedAt
}

// FlightDepartedEvent represents flight departure event
type FlightDepartedEvent struct {
	FlightID          string    `json:"flight_id"`
	FlightNumber      string    `json:"flight_number"`
	Origin            string    `json:"origin"`
	Destination       string    `json:"destination"`
	ActualDeparture   time.Time `json:"actual_departure_time"`
	ScheduledDeparture time.Time `json:"scheduled_departure_time"`
	DepartedAt        time.Time `json:"departed_at"`
}

func (e FlightDepartedEvent) EventType() string {
	return "flight.departed"
}

func (e FlightDepartedEvent) EventID() string {
	return e.FlightID
}

func (e FlightDepartedEvent) OccurredAt() time.Time {
	return e.DepartedAt
}

// FlightArrivedEvent represents flight arrival event
type FlightArrivedEvent struct {
	FlightID        string    `json:"flight_id"`
	FlightNumber    string    `json:"flight_number"`
	Origin          string    `json:"origin"`
	Destination     string    `json:"destination"`
	ActualArrival   time.Time `json:"actual_arrival_time"`
	ScheduledArrival time.Time `json:"scheduled_arrival_time"`
	ArrivedAt       time.Time `json:"arrived_at"`
}

func (e FlightArrivedEvent) EventType() string {
	return "flight.arrived"
}

func (e FlightArrivedEvent) EventID() string {
	return e.FlightID
}

func (e FlightArrivedEvent) OccurredAt() time.Time {
	return e.ArrivedAt
}

package entity

import (
	"testing"
	"time"

	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

func TestNewFlight(t *testing.T) {
	flightNumber, _ := valueobject.NewFlightNumber("AA123")
	origin := "JFK"
	destination := "LAX"
	departureTime := time.Now().Add(2 * time.Hour)
	arrivalTime := departureTime.Add(5 * time.Hour)
	aircraftType := "Boeing 737"
	gate := "A12"

	flight, err := NewFlight(flightNumber, origin, destination, departureTime, arrivalTime, aircraftType, gate)

	if err != nil {
		t.Fatalf("NewFlight() error = %v", err)
	}

	if flight.ID() == "" {
		t.Error("Flight ID should not be empty")
	}

	if !flight.FlightNumber().Equals(flightNumber) {
		t.Errorf("FlightNumber = %v, want %v", flight.FlightNumber(), flightNumber)
	}

	if flight.Origin() != origin {
		t.Errorf("Origin = %v, want %v", flight.Origin(), origin)
	}

	if !flight.Status().Equals(valueobject.Scheduled()) {
		t.Errorf("Status should be scheduled, got %v", flight.Status())
	}
}

func TestNewFlight_Validation(t *testing.T) {
	flightNumber, _ := valueobject.NewFlightNumber("AA123")
	now := time.Now()

	tests := []struct {
		name          string
		origin        string
		destination   string
		departureTime time.Time
		arrivalTime   time.Time
		wantErr       bool
	}{
		{
			name:          "valid flight",
			origin:        "JFK",
			destination:   "LAX",
			departureTime: now.Add(2 * time.Hour),
			arrivalTime:   now.Add(7 * time.Hour),
			wantErr:       false,
		},
		{
			name:          "empty origin",
			origin:        "",
			destination:   "LAX",
			departureTime: now.Add(2 * time.Hour),
			arrivalTime:   now.Add(7 * time.Hour),
			wantErr:       true,
		},
		{
			name:          "empty destination",
			origin:        "JFK",
			destination:   "",
			departureTime: now.Add(2 * time.Hour),
			arrivalTime:   now.Add(7 * time.Hour),
			wantErr:       true,
		},
		{
			name:          "same origin and destination",
			origin:        "JFK",
			destination:   "JFK",
			departureTime: now.Add(2 * time.Hour),
			arrivalTime:   now.Add(7 * time.Hour),
			wantErr:       true,
		},
		{
			name:          "arrival before departure",
			origin:        "JFK",
			destination:   "LAX",
			departureTime: now.Add(7 * time.Hour),
			arrivalTime:   now.Add(2 * time.Hour),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFlight(flightNumber, tt.origin, tt.destination, tt.departureTime, tt.arrivalTime, "Boeing 737", "A12")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFlight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlight_UpdateStatus(t *testing.T) {
	flight := createTestFlight(t)

	// Test valid transition: scheduled -> boarding
	boardingStatus := valueobject.Boarding()
	err := flight.UpdateStatus(boardingStatus)
	if err != nil {
		t.Errorf("UpdateStatus() error = %v", err)
	}
	if !flight.Status().Equals(boardingStatus) {
		t.Errorf("Status = %v, want %v", flight.Status(), boardingStatus)
	}

	// Test invalid transition: boarding -> arrived
	arrivedStatus := valueobject.Arrived()
	err = flight.UpdateStatus(arrivedStatus)
	if err == nil {
		t.Error("UpdateStatus() should error on invalid transition")
	}
}

func TestFlight_Delay(t *testing.T) {
	flight := createTestFlight(t)
	originalDeparture := flight.DepartureTime()
	delayDuration := 30 * time.Minute

	err := flight.Delay(delayDuration, "Weather conditions")
	if err != nil {
		t.Errorf("Delay() error = %v", err)
	}

	if !flight.Status().Equals(valueobject.Delayed()) {
		t.Errorf("Status should be delayed, got %v", flight.Status())
	}

	expectedDeparture := originalDeparture.Add(delayDuration)
	if !flight.DepartureTime().Equal(expectedDeparture) {
		t.Errorf("DepartureTime = %v, want %v", flight.DepartureTime(), expectedDeparture)
	}

	if flight.DelayReason() != "Weather conditions" {
		t.Errorf("DelayReason = %v, want 'Weather conditions'", flight.DelayReason())
	}
}

func TestFlight_Delay_AfterDeparted(t *testing.T) {
	flight := createTestFlight(t)

	// Transition to departed
	_ = flight.UpdateStatus(valueobject.Boarding())
	_ = flight.UpdateStatus(valueobject.Departed())

	// Try to delay after departed
	err := flight.Delay(30*time.Minute, "Too late")
	if err == nil {
		t.Error("Delay() should error after flight has departed")
	}
}

func TestFlight_Cancel(t *testing.T) {
	flight := createTestFlight(t)

	err := flight.Cancel("Maintenance required")
	if err != nil {
		t.Errorf("Cancel() error = %v", err)
	}

	if !flight.Status().Equals(valueobject.Cancelled()) {
		t.Errorf("Status should be cancelled, got %v", flight.Status())
	}

	if flight.CancellationReason() != "Maintenance required" {
		t.Errorf("CancellationReason = %v, want 'Maintenance required'", flight.CancellationReason())
	}
}

func TestFlight_Cancel_AfterDeparted(t *testing.T) {
	flight := createTestFlight(t)

	// Transition to in_flight
	_ = flight.UpdateStatus(valueobject.Boarding())
	_ = flight.UpdateStatus(valueobject.Departed())
	_ = flight.UpdateStatus(valueobject.InFlight())

	// Try to cancel after in flight
	err := flight.Cancel("Cannot cancel")
	if err == nil {
		t.Error("Cancel() should error after flight is in flight")
	}
}

func TestFlight_UpdateGate(t *testing.T) {
	flight := createTestFlight(t)

	err := flight.UpdateGate("B20")
	if err != nil {
		t.Errorf("UpdateGate() error = %v", err)
	}

	if flight.Gate() != "B20" {
		t.Errorf("Gate = %v, want B20", flight.Gate())
	}
}

func TestFlight_UpdateGate_AfterBoarding(t *testing.T) {
	flight := createTestFlight(t)
	_ = flight.UpdateStatus(valueobject.Boarding())

	err := flight.UpdateGate("B20")
	if err == nil {
		t.Error("UpdateGate() should error after boarding has started")
	}
}

func TestFlight_SetActualDepartureTime(t *testing.T) {
	flight := createTestFlight(t)
	_ = flight.UpdateStatus(valueobject.Boarding())

	actualTime := time.Now()
	err := flight.SetActualDepartureTime(actualTime)
	if err != nil {
		t.Errorf("SetActualDepartureTime() error = %v", err)
	}

	if !flight.ActualDepartureTime().Equal(actualTime) {
		t.Errorf("ActualDepartureTime = %v, want %v", flight.ActualDepartureTime(), actualTime)
	}
}

func TestFlight_SetActualArrivalTime(t *testing.T) {
	flight := createTestFlight(t)
	_ = flight.UpdateStatus(valueobject.Boarding())
	_ = flight.UpdateStatus(valueobject.Departed())
	_ = flight.UpdateStatus(valueobject.InFlight())
	_ = flight.UpdateStatus(valueobject.Landed())

	actualTime := time.Now()
	err := flight.SetActualArrivalTime(actualTime)
	if err != nil {
		t.Errorf("SetActualArrivalTime() error = %v", err)
	}

	if !flight.ActualArrivalTime().Equal(actualTime) {
		t.Errorf("ActualArrivalTime = %v, want %v", flight.ActualArrivalTime(), actualTime)
	}
}

func TestFlight_IsDelayed(t *testing.T) {
	flight := createTestFlight(t)

	if flight.IsDelayed() {
		t.Error("Flight should not be delayed initially")
	}

	_ = flight.Delay(30*time.Minute, "Weather")

	if !flight.IsDelayed() {
		t.Error("Flight should be delayed after Delay() is called")
	}
}

func TestFlight_IsActive(t *testing.T) {
	flight := createTestFlight(t)

	if !flight.IsActive() {
		t.Error("Scheduled flight should be active")
	}

	// Complete the flight
	_ = flight.UpdateStatus(valueobject.Boarding())
	_ = flight.UpdateStatus(valueobject.Departed())
	_ = flight.UpdateStatus(valueobject.InFlight())
	_ = flight.UpdateStatus(valueobject.Landed())
	_ = flight.UpdateStatus(valueobject.Arrived())

	if flight.IsActive() {
		t.Error("Arrived flight should not be active")
	}
}

// Helper function to create a test flight
func createTestFlight(t *testing.T) *Flight {
	t.Helper()
	flightNumber, _ := valueobject.NewFlightNumber("AA123")
	departureTime := time.Now().Add(2 * time.Hour)
	arrivalTime := departureTime.Add(5 * time.Hour)

	flight, err := NewFlight(
		flightNumber,
		"JFK",
		"LAX",
		departureTime,
		arrivalTime,
		"Boeing 737",
		"A12",
	)
	if err != nil {
		t.Fatalf("Failed to create test flight: %v", err)
	}
	return flight
}

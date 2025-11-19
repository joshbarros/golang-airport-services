package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/entity"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
)

// Mock repository for testing
type mockFlightRepository struct {
	saveFn                     func(ctx context.Context, flight *entity.Flight) error
	findByFlightNumberFn       func(ctx context.Context, flightNumber valueobject.FlightNumber, date time.Time) (*entity.Flight, error)
}

func (m *mockFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, flight)
	}
	return nil
}

func (m *mockFlightRepository) FindByID(ctx context.Context, id string) (*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindByFlightNumber(ctx context.Context, flightNumber valueobject.FlightNumber, date time.Time) (*entity.Flight, error) {
	if m.findByFlightNumberFn != nil {
		return m.findByFlightNumberFn(ctx, flightNumber, date)
	}
	return nil, nil
}

func (m *mockFlightRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindByStatus(ctx context.Context, status valueobject.FlightStatus, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindByOrigin(ctx context.Context, origin string, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindByDestination(ctx context.Context, destination string, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindByOriginAndDestination(ctx context.Context, origin, destination string, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindActiveFlights(ctx context.Context, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) FindDelayedFlights(ctx context.Context, limit, offset int) ([]*entity.Flight, error) {
	return nil, nil
}

func (m *mockFlightRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockFlightRepository) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockFlightRepository) CountByStatus(ctx context.Context, status valueobject.FlightStatus) (int64, error) {
	return 0, nil
}

var _ repository.FlightRepository = (*mockFlightRepository)(nil)

func TestCreateFlightUseCase_Execute(t *testing.T) {
	repo := &mockFlightRepository{
		saveFn: func(ctx context.Context, flight *entity.Flight) error {
			return nil
		},
		findByFlightNumberFn: func(ctx context.Context, flightNumber valueobject.FlightNumber, date time.Time) (*entity.Flight, error) {
			return nil, nil // No existing flight
		},
	}

	useCase := NewCreateFlightUseCase(repo)

	input := CreateFlightInput{
		FlightNumber:  "AA123",
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureTime: time.Now().Add(2 * time.Hour),
		ArrivalTime:   time.Now().Add(7 * time.Hour),
		AircraftType:  "Boeing 737",
		Gate:          "A12",
	}

	output, err := useCase.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if output.FlightNumber != "AA123" {
		t.Errorf("FlightNumber = %v, want AA123", output.FlightNumber)
	}

	if output.Status != "scheduled" {
		t.Errorf("Status = %v, want scheduled", output.Status)
	}
}

func TestCreateFlightUseCase_Execute_InvalidFlightNumber(t *testing.T) {
	repo := &mockFlightRepository{}
	useCase := NewCreateFlightUseCase(repo)

	input := CreateFlightInput{
		FlightNumber:  "INVALID",
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureTime: time.Now().Add(2 * time.Hour),
		ArrivalTime:   time.Now().Add(7 * time.Hour),
		AircraftType:  "Boeing 737",
		Gate:          "A12",
	}

	_, err := useCase.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for invalid flight number")
	}
}

func TestCreateFlightUseCase_Execute_DuplicateFlight(t *testing.T) {
	existingFlight, _ := createTestFlight()

	repo := &mockFlightRepository{
		findByFlightNumberFn: func(ctx context.Context, flightNumber valueobject.FlightNumber, date time.Time) (*entity.Flight, error) {
			return existingFlight, nil // Flight already exists
		},
	}

	useCase := NewCreateFlightUseCase(repo)

	input := CreateFlightInput{
		FlightNumber:  "AA123",
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureTime: time.Now().Add(2 * time.Hour),
		ArrivalTime:   time.Now().Add(7 * time.Hour),
		AircraftType:  "Boeing 737",
		Gate:          "A12",
	}

	_, err := useCase.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for duplicate flight")
	}
}

// Helper function to create a test flight
func createTestFlight() (*entity.Flight, error) {
	flightNumber, _ := valueobject.NewFlightNumber("AA123")
	return entity.NewFlight(
		flightNumber,
		"JFK",
		"LAX",
		time.Now().Add(2*time.Hour),
		time.Now().Add(7*time.Hour),
		"Boeing 737",
		"A12",
	)
}

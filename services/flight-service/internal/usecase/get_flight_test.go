package usecase

import (
	"context"
	"testing"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/entity"
)

func TestGetFlightUseCase_Execute(t *testing.T) {
	flight, _ := createTestFlight()

	repo := &mockFlightRepository{
		findByIDFn: func(ctx context.Context, id string) (*entity.Flight, error) {
			if id == flight.ID() {
				return flight, nil
			}
			return nil, errors.NotFound("flight not found")
		},
	}

	useCase := NewGetFlightUseCase(repo)

	output, err := useCase.Execute(context.Background(), flight.ID())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ID != flight.ID() {
		t.Errorf("ID = %v, want %v", output.ID, flight.ID())
	}

	if output.FlightNumber != flight.FlightNumber().String() {
		t.Errorf("FlightNumber = %v, want %v", output.FlightNumber, flight.FlightNumber().String())
	}
}

func TestGetFlightUseCase_Execute_NotFound(t *testing.T) {
	repo := &mockFlightRepository{
		findByIDFn: func(ctx context.Context, id string) (*entity.Flight, error) {
			return nil, errors.NotFound("flight not found")
		},
	}

	useCase := NewGetFlightUseCase(repo)

	_, err := useCase.Execute(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent flight")
	}

	if !errors.IsAppError(err) {
		t.Error("Expected AppError")
	}
}

func TestGetFlightUseCase_Execute_EmptyID(t *testing.T) {
	repo := &mockFlightRepository{}
	useCase := NewGetFlightUseCase(repo)

	_, err := useCase.Execute(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

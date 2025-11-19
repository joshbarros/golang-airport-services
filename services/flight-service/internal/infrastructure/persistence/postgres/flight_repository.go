package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/entity"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/valueobject"
	"go.uber.org/zap"
)

// flightRepository implements repository.FlightRepository using PostgreSQL
type flightRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewFlightRepository creates a new PostgreSQL flight repository
func NewFlightRepository(pool *pgxpool.Pool, log *logger.Logger) repository.FlightRepository {
	return &flightRepository{
		pool:   pool,
		logger: log,
	}
}

// Save persists a flight (create or update)
func (r *flightRepository) Save(ctx context.Context, flight *entity.Flight) error {
	query := `
		INSERT INTO flights (
			id, flight_number, origin, destination,
			departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			status, aircraft_type, gate, terminal,
			delay_reason, cancellation_reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		ON CONFLICT (id) DO UPDATE SET
			flight_number = EXCLUDED.flight_number,
			origin = EXCLUDED.origin,
			destination = EXCLUDED.destination,
			departure_time = EXCLUDED.departure_time,
			arrival_time = EXCLUDED.arrival_time,
			actual_departure_time = EXCLUDED.actual_departure_time,
			actual_arrival_time = EXCLUDED.actual_arrival_time,
			status = EXCLUDED.status,
			aircraft_type = EXCLUDED.aircraft_type,
			gate = EXCLUDED.gate,
			terminal = EXCLUDED.terminal,
			delay_reason = EXCLUDED.delay_reason,
			cancellation_reason = EXCLUDED.cancellation_reason,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		flight.ID(),
		flight.FlightNumber().String(),
		flight.Origin(),
		flight.Destination(),
		flight.DepartureTime(),
		flight.ArrivalTime(),
		nullableTime(flight.ActualDepartureTime()),
		nullableTime(flight.ActualArrivalTime()),
		flight.Status().String(),
		flight.AircraftType(),
		flight.Gate(),
		flight.Terminal(),
		flight.DelayReason(),
		flight.CancellationReason(),
		flight.CreatedAt(),
		flight.UpdatedAt(),
	)

	if err != nil {
		r.logger.Error("Failed to save flight",
			zap.String("flight_id", flight.ID()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save flight: %w", err)
	}

	r.logger.Debug("Flight saved successfully",
		zap.String("flight_id", flight.ID()),
		zap.String("flight_number", flight.FlightNumber().String()),
	)

	return nil
}

// FindByID retrieves a flight by its ID
func (r *flightRepository) FindByID(ctx context.Context, id string) (*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	return r.scanFlight(row)
}

// FindByFlightNumber retrieves a flight by flight number and date
func (r *flightRepository) FindByFlightNumber(ctx context.Context, flightNumber valueobject.FlightNumber, date time.Time) (*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE flight_number = $1
		  AND DATE(departure_time) = DATE($2)
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, flightNumber.String(), date)
	return r.scanFlight(row)
}

// FindAll retrieves all flights with pagination
func (r *flightRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		ORDER BY departure_time DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flights: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// FindByStatus retrieves flights by status with pagination
func (r *flightRepository) FindByStatus(ctx context.Context, status valueobject.FlightStatus, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE status = $1
		ORDER BY departure_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, status.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flights by status: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// FindByOrigin retrieves flights by origin airport
func (r *flightRepository) FindByOrigin(ctx context.Context, origin string, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE origin = $1
		ORDER BY departure_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, origin, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flights by origin: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// FindByDestination retrieves flights by destination airport
func (r *flightRepository) FindByDestination(ctx context.Context, destination string, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE destination = $1
		ORDER BY departure_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, destination, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flights by destination: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// FindByOriginAndDestination retrieves flights by route
func (r *flightRepository) FindByOriginAndDestination(ctx context.Context, origin, destination string, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE origin = $1 AND destination = $2
		ORDER BY departure_time DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, origin, destination, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flights by route: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// FindByDateRange retrieves flights within a date range
func (r *flightRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE departure_time >= $1 AND departure_time <= $2
		ORDER BY departure_time ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flights by date range: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// FindActiveFlights retrieves all active flights
func (r *flightRepository) FindActiveFlights(ctx context.Context, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE status NOT IN ('arrived', 'cancelled')
		ORDER BY departure_time ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query active flights: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// FindDelayedFlights retrieves all delayed flights
func (r *flightRepository) FindDelayedFlights(ctx context.Context, limit, offset int) ([]*entity.Flight, error) {
	query := `
		SELECT id, flight_number, origin, destination,
			   departure_time, arrival_time, actual_departure_time, actual_arrival_time,
			   status, aircraft_type, gate, terminal,
			   delay_reason, cancellation_reason, created_at, updated_at
		FROM flights
		WHERE status = 'delayed'
		ORDER BY departure_time ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query delayed flights: %w", err)
	}
	defer rows.Close()

	return r.scanFlights(rows)
}

// Delete removes a flight
func (r *flightRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM flights WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete flight: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("flight not found")
	}

	r.logger.Debug("Flight deleted", zap.String("flight_id", id))
	return nil
}

// Count returns total number of flights
func (r *flightRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM flights").Scan(&count)
	return count, err
}

// CountByStatus returns count of flights by status
func (r *flightRepository) CountByStatus(ctx context.Context, status valueobject.FlightStatus) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM flights WHERE status = $1"
	err := r.pool.QueryRow(ctx, query, status.String()).Scan(&count)
	return count, err
}

// Helper methods

// scanFlight scans a single row into a Flight entity
func (r *flightRepository) scanFlight(row pgx.Row) (*entity.Flight, error) {
	var (
		id                  string
		flightNumber        string
		origin              string
		destination         string
		departureTime       time.Time
		arrivalTime         time.Time
		actualDepartureTime *time.Time
		actualArrivalTime   *time.Time
		status              string
		aircraftType        string
		gate                string
		terminal            string
		delayReason         string
		cancellationReason  string
		createdAt           time.Time
		updatedAt           time.Time
	)

	err := row.Scan(
		&id, &flightNumber, &origin, &destination,
		&departureTime, &arrivalTime, &actualDepartureTime, &actualArrivalTime,
		&status, &aircraftType, &gate, &terminal,
		&delayReason, &cancellationReason, &createdAt, &updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan flight: %w", err)
	}

	// Reconstruct value objects
	flightNum, err := valueobject.NewFlightNumber(flightNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid flight number in database: %w", err)
	}

	flightStatus, err := valueobject.NewFlightStatus(status)
	if err != nil {
		return nil, fmt.Errorf("invalid flight status in database: %w", err)
	}

	// Reconstruct entity
	flight := entity.Reconstruct(
		id, flightNum, origin, destination,
		departureTime, arrivalTime,
		valueOrZero(actualDepartureTime),
		valueOrZero(actualArrivalTime),
		flightStatus,
		aircraftType, gate, terminal,
		delayReason, cancellationReason,
		createdAt, updatedAt,
	)

	return flight, nil
}

// scanFlights scans multiple rows into Flight entities
func (r *flightRepository) scanFlights(rows pgx.Rows) ([]*entity.Flight, error) {
	var flights []*entity.Flight

	for rows.Next() {
		var (
			id                  string
			flightNumber        string
			origin              string
			destination         string
			departureTime       time.Time
			arrivalTime         time.Time
			actualDepartureTime *time.Time
			actualArrivalTime   *time.Time
			status              string
			aircraftType        string
			gate                string
			terminal            string
			delayReason         string
			cancellationReason  string
			createdAt           time.Time
			updatedAt           time.Time
		)

		err := rows.Scan(
			&id, &flightNumber, &origin, &destination,
			&departureTime, &arrivalTime, &actualDepartureTime, &actualArrivalTime,
			&status, &aircraftType, &gate, &terminal,
			&delayReason, &cancellationReason, &createdAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan flight row: %w", err)
		}

		// Reconstruct value objects
		flightNum, err := valueobject.NewFlightNumber(flightNumber)
		if err != nil {
			r.logger.Warn("Invalid flight number in database",
				zap.String("flight_number", flightNumber),
				zap.Error(err),
			)
			continue
		}

		flightStatus, err := valueobject.NewFlightStatus(status)
		if err != nil {
			r.logger.Warn("Invalid flight status in database",
				zap.String("status", status),
				zap.Error(err),
			)
			continue
		}

		// Reconstruct entity
		flight := entity.Reconstruct(
			id, flightNum, origin, destination,
			departureTime, arrivalTime,
			valueOrZero(actualDepartureTime),
			valueOrZero(actualArrivalTime),
			flightStatus,
			aircraftType, gate, terminal,
			delayReason, cancellationReason,
			createdAt, updatedAt,
		)

		flights = append(flights, flight)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flight rows: %w", err)
	}

	return flights, nil
}

// Helper functions

func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func valueOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

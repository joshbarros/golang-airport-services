# Clean Architecture Implementation Guide

## Overview

This guide details how we implement Clean Architecture in the Airport Services microservices platform. Clean Architecture ensures our codebase is maintainable, testable, and independent of frameworks, databases, and external agencies.

## Core Principles

### 1. Independence
- **Framework Independence**: Business logic doesn't depend on frameworks
- **Database Independence**: Business logic doesn't know about database details
- **UI Independence**: Can change UI without affecting business logic
- **External Agency Independence**: Business rules don't depend on external services
- **Testability**: Business logic can be tested without external dependencies

### 2. The Dependency Rule

**Dependencies point inward**. Source code dependencies can only point inward toward higher-level policies.

```
┌─────────────────────────────────────────┐
│         External Interfaces             │  ← Frameworks, Drivers, Web, DB
│  ┌───────────────────────────────────┐  │
│  │    Interface Adapters             │  │  ← Controllers, Gateways, Presenters
│  │  ┌─────────────────────────────┐  │  │
│  │  │   Application Business       │  │  │  ← Use Cases, Interactors
│  │  │   Rules (Use Cases)          │  │  │
│  │  │  ┌───────────────────────┐   │  │  │
│  │  │  │  Enterprise Business  │   │  │  │  ← Entities, Domain Models
│  │  │  │  Rules (Entities)     │   │  │  │
│  │  │  │                       │   │  │  │
│  │  │  └───────────────────────┘   │  │  │
│  │  │                               │  │  │
│  │  └─────────────────────────────┘  │  │
│  │                                    │  │
│  └───────────────────────────────────┘  │
│                                          │
└─────────────────────────────────────────┘
```

## Layer Structure

### Layer 1: Entities (Domain Layer)

**Location**: `internal/{service}/domain/entity/`

**Purpose**: Enterprise-wide business rules and domain models

**Characteristics**:
- Pure business logic
- No external dependencies
- Framework agnostic
- Highly reusable

**Example**:
```go
// internal/flight/domain/entity/flight.go
package entity

import (
    "time"
    "errors"
)

// Flight represents the core business entity
type Flight struct {
    id            string
    flightNumber  string
    departure     Airport
    arrival       Airport
    departureTime time.Time
    arrivalTime   time.Time
    status        FlightStatus
    aircraft      *Aircraft
}

// NewFlight creates a new flight with business rules validation
func NewFlight(flightNumber string, departure, arrival Airport,
               departureTime, arrivalTime time.Time) (*Flight, error) {

    // Business Rule: Flight number must be valid format
    if !isValidFlightNumber(flightNumber) {
        return nil, errors.New("invalid flight number format")
    }

    // Business Rule: Arrival must be after departure
    if !arrivalTime.After(departureTime) {
        return nil, errors.New("arrival time must be after departure time")
    }

    // Business Rule: Departure and arrival airports must be different
    if departure.Code() == arrival.Code() {
        return nil, errors.New("departure and arrival airports must be different")
    }

    return &Flight{
        id:            generateID(),
        flightNumber:  flightNumber,
        departure:     departure,
        arrival:       arrival,
        departureTime: departureTime,
        arrivalTime:   arrivalTime,
        status:        FlightStatusScheduled,
    }, nil
}

// UpdateStatus changes flight status with business rules
func (f *Flight) UpdateStatus(newStatus FlightStatus) error {
    // Business Rule: Status transition validation
    if !f.canTransitionTo(newStatus) {
        return errors.New("invalid status transition")
    }

    f.status = newStatus
    return nil
}

// Delay adds delay to the flight
func (f *Flight) Delay(duration time.Duration, reason string) error {
    // Business Rule: Cannot delay already departed flights
    if f.status == FlightStatusDeparted || f.status == FlightStatusArrived {
        return errors.New("cannot delay flight that has already departed")
    }

    f.departureTime = f.departureTime.Add(duration)
    f.arrivalTime = f.arrivalTime.Add(duration)
    f.status = FlightStatusDelayed

    return nil
}

// canTransitionTo validates status transitions
func (f *Flight) canTransitionTo(newStatus FlightStatus) bool {
    validTransitions := map[FlightStatus][]FlightStatus{
        FlightStatusScheduled: {FlightStatusBoarding, FlightStatusDelayed, FlightStatusCancelled},
        FlightStatusDelayed:   {FlightStatusBoarding, FlightStatusCancelled},
        FlightStatusBoarding:  {FlightStatusDeparted, FlightStatusCancelled},
        FlightStatusDeparted:  {FlightStatusArrived, FlightStatusDiverted},
        FlightStatusDiverted:  {FlightStatusArrived},
    }

    allowed := validTransitions[f.status]
    for _, status := range allowed {
        if status == newStatus {
            return true
        }
    }
    return false
}

// Getters (read-only access to maintain encapsulation)
func (f *Flight) ID() string { return f.id }
func (f *Flight) FlightNumber() string { return f.flightNumber }
func (f *Flight) Status() FlightStatus { return f.status }
func (f *Flight) DepartureTime() time.Time { return f.departureTime }
func (f *Flight) ArrivalTime() time.Time { return f.arrivalTime }
```

**Key Points**:
- Rich domain models with behavior
- Encapsulation (private fields, public methods)
- Business rules enforced in the entity
- No dependencies on outer layers

---

### Layer 2: Use Cases (Application Layer)

**Location**: `internal/{service}/usecase/`

**Purpose**: Application-specific business rules and orchestration

**Characteristics**:
- Orchestrates the flow of data to/from entities
- Directs entities to use their business rules
- Coordinates between multiple entities
- Independent of delivery mechanism (HTTP, gRPC)

**Example**:
```go
// internal/flight/usecase/create_flight.go
package usecase

import (
    "context"
    "fmt"

    "github.com/joshbarros/golang-airport-services/internal/flight/domain/entity"
    "github.com/joshbarros/golang-airport-services/internal/flight/domain/repository"
    "github.com/joshbarros/golang-airport-services/internal/flight/domain/event"
)

// CreateFlightInput represents the input for creating a flight
type CreateFlightInput struct {
    FlightNumber     string
    DepartureCode    string
    ArrivalCode      string
    DepartureTime    time.Time
    ArrivalTime      time.Time
    AircraftID       string
}

// CreateFlightOutput represents the output
type CreateFlightOutput struct {
    FlightID     string
    FlightNumber string
    Status       string
}

// CreateFlightUseCase encapsulates the create flight business logic
type CreateFlightUseCase struct {
    flightRepo   repository.FlightRepository
    airportRepo  repository.AirportRepository
    aircraftRepo repository.AircraftRepository
    eventBus     event.EventBus
}

// NewCreateFlightUseCase creates a new use case
func NewCreateFlightUseCase(
    flightRepo repository.FlightRepository,
    airportRepo repository.AirportRepository,
    aircraftRepo repository.AircraftRepository,
    eventBus event.EventBus,
) *CreateFlightUseCase {
    return &CreateFlightUseCase{
        flightRepo:   flightRepo,
        airportRepo:  airportRepo,
        aircraftRepo: aircraftRepo,
        eventBus:     eventBus,
    }
}

// Execute runs the use case
func (uc *CreateFlightUseCase) Execute(ctx context.Context, input CreateFlightInput) (*CreateFlightOutput, error) {
    // 1. Validate and fetch airports
    departure, err := uc.airportRepo.FindByCode(ctx, input.DepartureCode)
    if err != nil {
        return nil, fmt.Errorf("invalid departure airport: %w", err)
    }

    arrival, err := uc.airportRepo.FindByCode(ctx, input.ArrivalCode)
    if err != nil {
        return nil, fmt.Errorf("invalid arrival airport: %w", err)
    }

    // 2. Validate aircraft availability
    aircraft, err := uc.aircraftRepo.FindByID(ctx, input.AircraftID)
    if err != nil {
        return nil, fmt.Errorf("aircraft not found: %w", err)
    }

    if !aircraft.IsAvailable(input.DepartureTime, input.ArrivalTime) {
        return nil, errors.New("aircraft not available for selected time")
    }

    // 3. Create flight entity (business rules applied here)
    flight, err := entity.NewFlight(
        input.FlightNumber,
        departure,
        arrival,
        input.DepartureTime,
        input.ArrivalTime,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create flight: %w", err)
    }

    // 4. Assign aircraft
    if err := flight.AssignAircraft(aircraft); err != nil {
        return nil, fmt.Errorf("failed to assign aircraft: %w", err)
    }

    // 5. Persist flight
    if err := uc.flightRepo.Save(ctx, flight); err != nil {
        return nil, fmt.Errorf("failed to save flight: %w", err)
    }

    // 6. Publish domain event
    uc.eventBus.Publish(ctx, event.FlightCreated{
        FlightID:     flight.ID(),
        FlightNumber: flight.FlightNumber(),
        DepartureTime: flight.DepartureTime(),
        ArrivalTime:   flight.ArrivalTime(),
    })

    // 7. Return output
    return &CreateFlightOutput{
        FlightID:     flight.ID(),
        FlightNumber: flight.FlightNumber(),
        Status:       flight.Status().String(),
    }, nil
}
```

**Key Points**:
- Uses interfaces for repositories (dependency inversion)
- Orchestrates multiple domain objects
- Application-specific validation
- Publishes domain events
- Returns DTOs, not domain entities

---

### Layer 3: Interface Adapters

**Location**: `internal/{service}/adapter/`

**Purpose**: Convert data between use cases and external systems

#### 3A. Controllers (Handlers)

**Location**: `internal/{service}/adapter/http/` or `internal/{service}/adapter/grpc/`

```go
// internal/flight/adapter/http/create_flight_handler.go
package http

import (
    "net/http"
    "github.com/gin-gonic/gin"

    "github.com/joshbarros/golang-airport-services/internal/flight/usecase"
)

// CreateFlightRequest represents the HTTP request
type CreateFlightRequest struct {
    FlightNumber  string `json:"flight_number" binding:"required"`
    Departure     string `json:"departure" binding:"required,len=3"`
    Arrival       string `json:"arrival" binding:"required,len=3"`
    DepartureTime string `json:"departure_time" binding:"required"`
    ArrivalTime   string `json:"arrival_time" binding:"required"`
    AircraftID    string `json:"aircraft_id" binding:"required,uuid"`
}

// CreateFlightResponse represents the HTTP response
type CreateFlightResponse struct {
    FlightID     string `json:"flight_id"`
    FlightNumber string `json:"flight_number"`
    Status       string `json:"status"`
}

// FlightHandler handles flight-related HTTP requests
type FlightHandler struct {
    createFlightUC *usecase.CreateFlightUseCase
}

// NewFlightHandler creates a new handler
func NewFlightHandler(createFlightUC *usecase.CreateFlightUseCase) *FlightHandler {
    return &FlightHandler{
        createFlightUC: createFlightUC,
    }
}

// CreateFlight handles POST /api/v1/flights
func (h *FlightHandler) CreateFlight(c *gin.Context) {
    var req CreateFlightRequest

    // 1. Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 2. Convert HTTP request to use case input
    input := h.toUseCaseInput(req)

    // 3. Execute use case
    output, err := h.createFlightUC.Execute(c.Request.Context(), input)
    if err != nil {
        // Handle different error types
        statusCode := h.mapErrorToHTTPStatus(err)
        c.JSON(statusCode, gin.H{"error": err.Error()})
        return
    }

    // 4. Convert use case output to HTTP response
    response := h.toHTTPResponse(output)

    // 5. Return response
    c.JSON(http.StatusCreated, response)
}

// toUseCaseInput converts HTTP request to use case input
func (h *FlightHandler) toUseCaseInput(req CreateFlightRequest) usecase.CreateFlightInput {
    departureTime, _ := time.Parse(time.RFC3339, req.DepartureTime)
    arrivalTime, _ := time.Parse(time.RFC3339, req.ArrivalTime)

    return usecase.CreateFlightInput{
        FlightNumber:  req.FlightNumber,
        DepartureCode: req.Departure,
        ArrivalCode:   req.Arrival,
        DepartureTime: departureTime,
        ArrivalTime:   arrivalTime,
        AircraftID:    req.AircraftID,
    }
}

// toHTTPResponse converts use case output to HTTP response
func (h *FlightHandler) toHTTPResponse(output *usecase.CreateFlightOutput) CreateFlightResponse {
    return CreateFlightResponse{
        FlightID:     output.FlightID,
        FlightNumber: output.FlightNumber,
        Status:       output.Status,
    }
}

// mapErrorToHTTPStatus maps domain errors to HTTP status codes
func (h *FlightHandler) mapErrorToHTTPStatus(err error) int {
    switch {
    case errors.Is(err, entity.ErrInvalidInput):
        return http.StatusBadRequest
    case errors.Is(err, entity.ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, entity.ErrConflict):
        return http.StatusConflict
    default:
        return http.StatusInternalServerError
    }
}
```

#### 3B. Gateways (Repository Implementations)

**Location**: `internal/{service}/adapter/repository/`

```go
// internal/flight/adapter/repository/postgres_flight_repository.go
package repository

import (
    "context"
    "database/sql"

    "github.com/joshbarros/golang-airport-services/internal/flight/domain/entity"
    "github.com/joshbarros/golang-airport-services/internal/flight/domain/repository"
)

// PostgresFlightRepository implements FlightRepository using PostgreSQL
type PostgresFlightRepository struct {
    db *sql.DB
}

// NewPostgresFlightRepository creates a new repository
func NewPostgresFlightRepository(db *sql.DB) repository.FlightRepository {
    return &PostgresFlightRepository{db: db}
}

// Save persists a flight
func (r *PostgresFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    query := `
        INSERT INTO flights (id, flight_number, departure_code, arrival_code,
                           departure_time, arrival_time, status, aircraft_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

    _, err := r.db.ExecContext(ctx, query,
        flight.ID(),
        flight.FlightNumber(),
        flight.Departure().Code(),
        flight.Arrival().Code(),
        flight.DepartureTime(),
        flight.ArrivalTime(),
        flight.Status().String(),
        flight.Aircraft().ID(),
    )

    return err
}

// FindByID retrieves a flight by ID
func (r *PostgresFlightRepository) FindByID(ctx context.Context, id string) (*entity.Flight, error) {
    query := `
        SELECT id, flight_number, departure_code, arrival_code,
               departure_time, arrival_time, status, aircraft_id
        FROM flights
        WHERE id = $1
    `

    var (
        flightID      string
        flightNumber  string
        departureCode string
        arrivalCode   string
        departureTime time.Time
        arrivalTime   time.Time
        status        string
        aircraftID    string
    )

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &flightID, &flightNumber, &departureCode, &arrivalCode,
        &departureTime, &arrivalTime, &status, &aircraftID,
    )

    if err == sql.ErrNoRows {
        return nil, entity.ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    // Reconstruct domain entity from database data
    return r.toDomain(flightID, flightNumber, departureCode, arrivalCode,
                      departureTime, arrivalTime, status, aircraftID)
}

// toDomain converts database model to domain entity
func (r *PostgresFlightRepository) toDomain(
    id, flightNumber, departureCode, arrivalCode string,
    departureTime, arrivalTime time.Time,
    status, aircraftID string,
) (*entity.Flight, error) {
    // Fetch related entities
    departure, _ := r.fetchAirport(departureCode)
    arrival, _ := r.fetchAirport(arrivalCode)
    aircraft, _ := r.fetchAircraft(aircraftID)

    // Use factory method to reconstruct entity
    return entity.ReconstructFlight(
        id, flightNumber, departure, arrival,
        departureTime, arrivalTime, entity.ParseStatus(status), aircraft,
    ), nil
}
```

#### 3C. Presenters

**Location**: `internal/{service}/adapter/presenter/`

```go
// internal/flight/adapter/presenter/flight_presenter.go
package presenter

import (
    "github.com/joshbarros/golang-airport-services/internal/flight/usecase"
)

// FlightPresenter formats use case output for display
type FlightPresenter struct{}

// PresentFlight formats flight data
func (p *FlightPresenter) PresentFlight(output *usecase.CreateFlightOutput) interface{} {
    return map[string]interface{}{
        "flight": map[string]string{
            "id":     output.FlightID,
            "number": output.FlightNumber,
            "status": output.Status,
        },
        "message": "Flight created successfully",
    }
}
```

---

### Layer 4: Frameworks & Drivers

**Location**: `cmd/{service}/`, `pkg/`

**Purpose**: External concerns (web framework, database driver, etc.)

```go
// cmd/flight-service/main.go
package main

import (
    "database/sql"
    "log"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"

    httpAdapter "github.com/joshbarros/golang-airport-services/internal/flight/adapter/http"
    repoAdapter "github.com/joshbarros/golang-airport-services/internal/flight/adapter/repository"
    "github.com/joshbarros/golang-airport-services/internal/flight/usecase"
    "github.com/joshbarros/golang-airport-services/pkg/config"
    "github.com/joshbarros/golang-airport-services/pkg/database"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize database
    db, err := database.NewPostgresConnection(cfg.Database)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Initialize repositories (Interface Adapters)
    flightRepo := repoAdapter.NewPostgresFlightRepository(db)
    airportRepo := repoAdapter.NewPostgresAirportRepository(db)
    aircraftRepo := repoAdapter.NewPostgresAircraftRepository(db)

    // Initialize event bus
    eventBus := initEventBus(cfg)

    // Initialize use cases (Application Layer)
    createFlightUC := usecase.NewCreateFlightUseCase(
        flightRepo,
        airportRepo,
        aircraftRepo,
        eventBus,
    )

    // Initialize handlers (Interface Adapters)
    flightHandler := httpAdapter.NewFlightHandler(createFlightUC)

    // Initialize router (Frameworks & Drivers)
    router := gin.Default()

    // Register routes
    v1 := router.Group("/api/v1")
    {
        v1.POST("/flights", flightHandler.CreateFlight)
    }

    // Start server
    log.Fatal(router.Run(":8081"))
}
```

---

## Directory Structure

```
internal/
└── flight/
    ├── domain/                    # Layer 1: Entities
    │   ├── entity/
    │   │   ├── flight.go         # Core business entities
    │   │   ├── aircraft.go
    │   │   ├── airport.go
    │   │   └── flight_status.go
    │   ├── repository/            # Repository interfaces (ports)
    │   │   ├── flight_repository.go
    │   │   ├── aircraft_repository.go
    │   │   └── airport_repository.go
    │   ├── event/                 # Domain events
    │   │   ├── flight_created.go
    │   │   └── flight_status_changed.go
    │   └── service/               # Domain services
    │       └── flight_validator.go
    │
    ├── usecase/                   # Layer 2: Use Cases
    │   ├── create_flight.go
    │   ├── update_flight_status.go
    │   ├── cancel_flight.go
    │   └── get_flight.go
    │
    ├── adapter/                   # Layer 3: Interface Adapters
    │   ├── http/                  # HTTP controllers
    │   │   ├── flight_handler.go
    │   │   ├── dto/               # Data Transfer Objects
    │   │   │   ├── request.go
    │   │   │   └── response.go
    │   │   └── middleware/
    │   ├── grpc/                  # gRPC controllers
    │   │   └── flight_service.go
    │   ├── repository/            # Repository implementations
    │   │   ├── postgres_flight_repository.go
    │   │   ├── postgres_aircraft_repository.go
    │   │   └── model/             # Database models (if using ORM)
    │   ├── event/                 # Event publishers/subscribers
    │   │   ├── rabbitmq_publisher.go
    │   │   └── rabbitmq_subscriber.go
    │   └── presenter/             # Response formatters
    │       └── flight_presenter.go
    │
    └── infrastructure/            # Layer 4: Frameworks & Drivers (if needed)
        └── cache/
            └── redis_cache.go
```

---

## Dependency Injection

Use constructor injection to maintain dependency inversion:

```go
// internal/flight/adapter/http/handler.go
type FlightHandler struct {
    createFlightUC usecase.CreateFlightUseCase    // Depends on use case interface
    updateFlightUC usecase.UpdateFlightUseCase
    getFlightUC    usecase.GetFlightUseCase
}

func NewFlightHandler(
    createFlightUC usecase.CreateFlightUseCase,
    updateFlightUC usecase.UpdateFlightUseCase,
    getFlightUC usecase.GetFlightUseCase,
) *FlightHandler {
    return &FlightHandler{
        createFlightUC: createFlightUC,
        updateFlightUC: updateFlightUC,
        getFlightUC:    getFlightUC,
    }
}
```

---

## Interface Segregation

Define small, focused interfaces:

```go
// domain/repository/flight_repository.go
package repository

// FlightRepository defines persistence operations
type FlightRepository interface {
    Save(ctx context.Context, flight *entity.Flight) error
    FindByID(ctx context.Context, id string) (*entity.Flight, error)
    FindByFlightNumber(ctx context.Context, number string) (*entity.Flight, error)
    Delete(ctx context.Context, id string) error
}

// FlightQueryRepository defines read operations (CQRS)
type FlightQueryRepository interface {
    Search(ctx context.Context, criteria SearchCriteria) ([]*FlightDTO, error)
    GetDepartureBoard(ctx context.Context, airportCode string) ([]*FlightDTO, error)
}
```

---

## Testing Strategy

### Entity Tests (Unit Tests)
```go
func TestFlight_UpdateStatus_ValidTransition(t *testing.T) {
    // Arrange
    flight := createTestFlight()

    // Act
    err := flight.UpdateStatus(entity.FlightStatusBoarding)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, entity.FlightStatusBoarding, flight.Status())
}
```

### Use Case Tests (Unit Tests with Mocks)
```go
func TestCreateFlightUseCase_Execute_Success(t *testing.T) {
    // Arrange
    mockFlightRepo := new(MockFlightRepository)
    mockAirportRepo := new(MockAirportRepository)
    mockEventBus := new(MockEventBus)

    uc := usecase.NewCreateFlightUseCase(
        mockFlightRepo,
        mockAirportRepo,
        mockAircraftRepo,
        mockEventBus,
    )

    mockAirportRepo.On("FindByCode", mock.Anything, "JFK").
        Return(jfkAirport, nil)
    mockFlightRepo.On("Save", mock.Anything, mock.Anything).
        Return(nil)

    input := usecase.CreateFlightInput{
        FlightNumber:  "AA123",
        DepartureCode: "JFK",
        // ...
    }

    // Act
    output, err := uc.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, output)
    mockFlightRepo.AssertExpectations(t)
}
```

### Handler Tests (Integration Tests)
```go
func TestFlightHandler_CreateFlight(t *testing.T) {
    // Arrange
    handler := setupTestHandler(t)

    reqBody := `{
        "flight_number": "AA123",
        "departure": "JFK",
        "arrival": "LAX"
    }`

    req := httptest.NewRequest("POST", "/api/v1/flights",
                               strings.NewReader(reqBody))
    w := httptest.NewRecorder()

    // Act
    handler.CreateFlight(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

---

## Benefits

1. **Testability**: Each layer can be tested independently
2. **Flexibility**: Easy to swap implementations (e.g., Postgres → MongoDB)
3. **Maintainability**: Clear separation of concerns
4. **Independence**: Business logic isolated from frameworks
5. **Scalability**: Easy to add new features without affecting existing code

---

## Common Pitfalls to Avoid

❌ **Don't let domain entities depend on database models**
```go
// BAD
type Flight struct {
    gorm.Model
    FlightNumber string
}
```

✅ **Keep domain entities pure**
```go
// GOOD
type Flight struct {
    id           string
    flightNumber string
}
```

❌ **Don't return domain entities from handlers**
```go
// BAD
func (h *Handler) GetFlight(c *gin.Context) {
    flight, _ := h.flightRepo.FindByID(id)
    c.JSON(200, flight)  // Exposes internal structure
}
```

✅ **Use DTOs for external communication**
```go
// GOOD
func (h *Handler) GetFlight(c *gin.Context) {
    output, _ := h.getFlightUC.Execute(id)
    response := toDTO(output)
    c.JSON(200, response)
}
```

❌ **Don't put business logic in handlers**
```go
// BAD
func (h *Handler) CreateFlight(c *gin.Context) {
    flight := &Flight{}
    if flight.ArrivalTime.Before(flight.DepartureTime) {
        c.JSON(400, "Invalid times")
        return
    }
}
```

✅ **Put business logic in entities/use cases**
```go
// GOOD
func (h *Handler) CreateFlight(c *gin.Context) {
    output, err := h.createFlightUC.Execute(input)
    if err != nil {
        c.JSON(400, err.Error())
        return
    }
}
```

---

## Summary

Clean Architecture provides:
- Clear boundaries between layers
- Dependency inversion (dependencies point inward)
- Testability at every level
- Framework independence
- Database independence
- Business logic isolation

By following these principles, our codebase remains maintainable, testable, and adaptable to changing requirements.

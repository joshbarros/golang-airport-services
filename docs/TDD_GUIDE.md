# Test-Driven Development (TDD) Guide

## Overview

Test-Driven Development is our development methodology where tests are written before production code. This guide explains how to practice TDD in the Airport Services platform.

## The TDD Cycle: Red-Green-Refactor

```
┌─────────────────────────────────────────┐
│                                         │
│  1. RED: Write a failing test           │
│     ↓                                   │
│  2. GREEN: Write minimal code to pass   │
│     ↓                                   │
│  3. REFACTOR: Improve code quality      │
│     ↓                                   │
│  4. Repeat                              │
│                                         │
└─────────────────────────────────────────┘
```

### Rules of TDD

1. **Write NO production code without a failing test**
2. **Write only enough test to fail** (compilation failures count as failures)
3. **Write only enough production code to make the failing test pass**

---

## Step-by-Step TDD Example

Let's build a `FlightNumber` value object using TDD.

### Step 1: RED - Write the First Failing Test

```go
// internal/flight/domain/value/flight_number_test.go
package value_test

import (
    "testing"
    "github.com/stretchr/testify/assert"

    "github.com/joshbarros/golang-airport-services/internal/flight/domain/value"
)

func TestFlightNumber_NewFlightNumber_ValidFormat(t *testing.T) {
    // Arrange
    input := "AA123"

    // Act
    flightNumber, err := value.NewFlightNumber(input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "AA123", flightNumber.String())
}
```

**Run Test**: `go test` → ❌ **FAILS** (compilation error: package/function doesn't exist)

### Step 2: GREEN - Write Minimal Code to Pass

```go
// internal/flight/domain/value/flight_number.go
package value

type FlightNumber struct {
    code string
}

func NewFlightNumber(code string) (FlightNumber, error) {
    return FlightNumber{code: code}, nil
}

func (fn FlightNumber) String() string {
    return fn.code
}
```

**Run Test**: `go test` → ✅ **PASSES**

### Step 3: REFACTOR - Improve (if needed)

Code is simple enough, no refactoring needed yet.

### Step 4: Add Next Test (RED)

```go
func TestFlightNumber_NewFlightNumber_InvalidFormat_TooShort(t *testing.T) {
    // Arrange
    input := "A1"

    // Act
    _, err := value.NewFlightNumber(input)

    // Assert
    assert.Error(t, err)
    assert.Equal(t, value.ErrInvalidFlightNumber, err)
}
```

**Run Test**: `go test` → ❌ **FAILS** (no validation)

### Step 5: GREEN - Add Validation

```go
// internal/flight/domain/value/flight_number.go
package value

import "errors"

var ErrInvalidFlightNumber = errors.New("invalid flight number format")

type FlightNumber struct {
    code string
}

func NewFlightNumber(code string) (FlightNumber, error) {
    if len(code) < 3 {
        return FlightNumber{}, ErrInvalidFlightNumber
    }
    return FlightNumber{code: code}, nil
}

func (fn FlightNumber) String() string {
    return fn.code
}
```

**Run Test**: `go test` → ✅ **PASSES**

### Step 6: Add More Tests and Iterate

```go
func TestFlightNumber_NewFlightNumber_InvalidFormat_NoAirlineCode(t *testing.T) {
    input := "123"
    _, err := value.NewFlightNumber(input)
    assert.Error(t, err)
}

func TestFlightNumber_NewFlightNumber_ValidFormat_WithSuffix(t *testing.T) {
    input := "BA2490A"
    flightNumber, err := value.NewFlightNumber(input)
    assert.NoError(t, err)
    assert.Equal(t, "BA2490A", flightNumber.String())
}
```

Continue the cycle: **RED** → **GREEN** → **REFACTOR** → **REPEAT**

---

## Complete TDD Example: Create Flight Use Case

### Iteration 1: Basic Creation

#### RED: Write Test

```go
// internal/flight/usecase/create_flight_test.go
package usecase_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/joshbarros/golang-airport-services/internal/flight/usecase"
)

func TestCreateFlightUseCase_Execute_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockFlightRepository)
    uc := usecase.NewCreateFlightUseCase(mockRepo)

    input := usecase.CreateFlightInput{
        FlightNumber:  "AA100",
        DepartureCode: "JFK",
        ArrivalCode:   "LAX",
        DepartureTime: time.Now().Add(24 * time.Hour),
        ArrivalTime:   time.Now().Add(30 * time.Hour),
    }

    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

    // Act
    output, err := uc.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, output)
    assert.Equal(t, "AA100", output.FlightNumber)
    mockRepo.AssertExpectations(t)
}
```

**Run**: ❌ **FAILS** (compilation error)

#### GREEN: Implement Minimal Code

```go
// internal/flight/usecase/create_flight.go
package usecase

import (
    "context"
    "time"

    "github.com/joshbarros/golang-airport-services/internal/flight/domain/entity"
    "github.com/joshbarros/golang-airport-services/internal/flight/domain/repository"
)

type CreateFlightInput struct {
    FlightNumber  string
    DepartureCode string
    ArrivalCode   string
    DepartureTime time.Time
    ArrivalTime   time.Time
}

type CreateFlightOutput struct {
    FlightID     string
    FlightNumber string
}

type CreateFlightUseCase struct {
    flightRepo repository.FlightRepository
}

func NewCreateFlightUseCase(repo repository.FlightRepository) *CreateFlightUseCase {
    return &CreateFlightUseCase{flightRepo: repo}
}

func (uc *CreateFlightUseCase) Execute(ctx context.Context, input CreateFlightInput) (*CreateFlightOutput, error) {
    flight := entity.NewFlight(input.FlightNumber, input.DepartureCode, input.ArrivalCode)

    if err := uc.flightRepo.Save(ctx, flight); err != nil {
        return nil, err
    }

    return &CreateFlightOutput{
        FlightID:     flight.ID(),
        FlightNumber: flight.FlightNumber(),
    }, nil
}
```

**Run**: ✅ **PASSES**

### Iteration 2: Validation

#### RED: Test Invalid Input

```go
func TestCreateFlightUseCase_Execute_InvalidFlightNumber(t *testing.T) {
    // Arrange
    mockRepo := new(MockFlightRepository)
    uc := usecase.NewCreateFlightUseCase(mockRepo)

    input := usecase.CreateFlightInput{
        FlightNumber:  "123", // Invalid
        DepartureCode: "JFK",
        ArrivalCode:   "LAX",
        DepartureTime: time.Now().Add(24 * time.Hour),
        ArrivalTime:   time.Now().Add(30 * time.Hour),
    }

    // Act
    _, err := uc.Execute(context.Background(), input)

    // Assert
    assert.Error(t, err)
    assert.Equal(t, entity.ErrInvalidFlightNumber, err)
    mockRepo.AssertNotCalled(t, "Save")
}
```

**Run**: ❌ **FAILS**

#### GREEN: Add Validation

```go
func (uc *CreateFlightUseCase) Execute(ctx context.Context, input CreateFlightInput) (*CreateFlightOutput, error) {
    // Validate flight number
    flightNumber, err := value.NewFlightNumber(input.FlightNumber)
    if err != nil {
        return nil, err
    }

    flight := entity.NewFlight(flightNumber, input.DepartureCode, input.ArrivalCode)

    if err := uc.flightRepo.Save(ctx, flight); err != nil {
        return nil, err
    }

    return &CreateFlightOutput{
        FlightID:     flight.ID(),
        FlightNumber: flight.FlightNumber(),
    }, nil
}
```

**Run**: ✅ **PASSES**

### Iteration 3: Business Rule Validation

#### RED: Test Times Validation

```go
func TestCreateFlightUseCase_Execute_ArrivalBeforeDeparture(t *testing.T) {
    // Arrange
    mockRepo := new(MockFlightRepository)
    uc := usecase.NewCreateFlightUseCase(mockRepo)

    now := time.Now()
    input := usecase.CreateFlightInput{
        FlightNumber:  "AA100",
        DepartureCode: "JFK",
        ArrivalCode:   "LAX",
        DepartureTime: now.Add(24 * time.Hour),
        ArrivalTime:   now.Add(20 * time.Hour), // Before departure!
    }

    // Act
    _, err := uc.Execute(context.Background(), input)

    // Assert
    assert.Error(t, err)
    assert.Equal(t, entity.ErrInvalidFlightTimes, err)
}
```

**Run**: ❌ **FAILS**

#### GREEN: Add Business Rule

```go
func (uc *CreateFlightUseCase) Execute(ctx context.Context, input CreateFlightInput) (*CreateFlightOutput, error) {
    // Validate flight number
    flightNumber, err := value.NewFlightNumber(input.FlightNumber)
    if err != nil {
        return nil, err
    }

    // Business rule: arrival must be after departure
    if !input.ArrivalTime.After(input.DepartureTime) {
        return nil, entity.ErrInvalidFlightTimes
    }

    flight := entity.NewFlight(
        flightNumber,
        input.DepartureCode,
        input.ArrivalCode,
        input.DepartureTime,
        input.ArrivalTime,
    )

    if err := uc.flightRepo.Save(ctx, flight); err != nil {
        return nil, err
    }

    return &CreateFlightOutput{
        FlightID:     flight.ID(),
        FlightNumber: flight.FlightNumber(),
    }, nil
}
```

**Run**: ✅ **PASSES**

#### REFACTOR: Extract Validation

```go
func (uc *CreateFlightUseCase) Execute(ctx context.Context, input CreateFlightInput) (*CreateFlightOutput, error) {
    // Validate input
    if err := uc.validateInput(input); err != nil {
        return nil, err
    }

    // Create flight entity
    flight, err := uc.createFlight(input)
    if err != nil {
        return nil, err
    }

    // Persist
    if err := uc.flightRepo.Save(ctx, flight); err != nil {
        return nil, err
    }

    return uc.buildOutput(flight), nil
}

func (uc *CreateFlightUseCase) validateInput(input CreateFlightInput) error {
    if _, err := value.NewFlightNumber(input.FlightNumber); err != nil {
        return err
    }

    if !input.ArrivalTime.After(input.DepartureTime) {
        return entity.ErrInvalidFlightTimes
    }

    return nil
}

func (uc *CreateFlightUseCase) createFlight(input CreateFlightInput) (*entity.Flight, error) {
    return entity.NewFlight(
        input.FlightNumber,
        input.DepartureCode,
        input.ArrivalCode,
        input.DepartureTime,
        input.ArrivalTime,
    )
}

func (uc *CreateFlightUseCase) buildOutput(flight *entity.Flight) *CreateFlightOutput {
    return &CreateFlightOutput{
        FlightID:     flight.ID(),
        FlightNumber: flight.FlightNumber(),
    }
}
```

**Run All Tests**: ✅ **ALL PASS**

---

## TDD Test Patterns

### Table-Driven Tests

Efficient way to test multiple scenarios:

```go
func TestFlightNumber_Validation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errType error
    }{
        {
            name:    "valid format",
            input:   "AA123",
            wantErr: false,
        },
        {
            name:    "valid with suffix",
            input:   "BA2490A",
            wantErr: false,
        },
        {
            name:    "too short",
            input:   "A1",
            wantErr: true,
            errType: value.ErrInvalidFlightNumber,
        },
        {
            name:    "no airline code",
            input:   "123",
            wantErr: true,
            errType: value.ErrInvalidFlightNumber,
        },
        {
            name:    "empty string",
            input:   "",
            wantErr: true,
            errType: value.ErrInvalidFlightNumber,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Act
            _, err := value.NewFlightNumber(tt.input)

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.Equal(t, tt.errType, err)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Test Fixtures and Builders

Create reusable test data:

```go
// test/fixtures/flight_builder.go
package fixtures

type FlightBuilder struct {
    flightNumber  string
    departureCode string
    arrivalCode   string
    departureTime time.Time
    arrivalTime   time.Time
}

func NewFlightBuilder() *FlightBuilder {
    now := time.Now()
    return &FlightBuilder{
        flightNumber:  "AA100",
        departureCode: "JFK",
        arrivalCode:   "LAX",
        departureTime: now.Add(24 * time.Hour),
        arrivalTime:   now.Add(30 * time.Hour),
    }
}

func (b *FlightBuilder) WithFlightNumber(number string) *FlightBuilder {
    b.flightNumber = number
    return b
}

func (b *FlightBuilder) WithDeparture(code string, time time.Time) *FlightBuilder {
    b.departureCode = code
    b.departureTime = time
    return b
}

func (b *FlightBuilder) WithArrival(code string, time time.Time) *FlightBuilder {
    b.arrivalCode = code
    b.arrivalTime = time
    return b
}

func (b *FlightBuilder) Build() *entity.Flight {
    flight, _ := entity.NewFlight(
        b.flightNumber,
        b.departureCode,
        b.arrivalCode,
        b.departureTime,
        b.arrivalTime,
    )
    return flight
}

// Usage in tests
func TestSomething(t *testing.T) {
    flight := fixtures.NewFlightBuilder().
        WithFlightNumber("BA456").
        WithDeparture("LHR", time.Now()).
        Build()

    // Test with flight
}
```

### Mocking with testify/mock

```go
// test/mocks/flight_repository_mock.go
package mocks

import (
    "context"
    "github.com/stretchr/testify/mock"

    "github.com/joshbarros/golang-airport-services/internal/flight/domain/entity"
)

type MockFlightRepository struct {
    mock.Mock
}

func (m *MockFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    args := m.Called(ctx, flight)
    return args.Error(0)
}

func (m *MockFlightRepository) FindByID(ctx context.Context, id string) (*entity.Flight, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.Flight), args.Error(1)
}

// Usage
func TestWithMock(t *testing.T) {
    mockRepo := new(mocks.MockFlightRepository)

    // Setup expectations
    mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(f *entity.Flight) bool {
        return f.FlightNumber() == "AA100"
    })).Return(nil)

    // Run test
    // ...

    // Verify
    mockRepo.AssertExpectations(t)
    mockRepo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}
```

---

## Testing Pyramid

```
        ┌──────────┐
        │   E2E    │  ← Few: Complete user flows
        ├──────────┤
        │Integration│ ← Some: Service interactions
        ├──────────┤
        │   Unit    │  ← Many: Individual components
        └──────────┘
```

### Unit Tests (70%)

Test individual components in isolation.

**What to test:**
- Entities and value objects
- Domain services
- Use cases
- Business logic

**Example:**
```go
func TestFlight_Delay_Success(t *testing.T) {
    // Arrange
    flight := fixtures.NewFlightBuilder().Build()
    delay := 2 * time.Hour

    // Act
    err := flight.Delay(delay, "Weather")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, entity.FlightStatusDelayed, flight.Status())
}

func TestFlight_Delay_AlreadyDeparted(t *testing.T) {
    // Arrange
    flight := fixtures.NewFlightBuilder().Build()
    flight.MarkAsDeparted()
    delay := 2 * time.Hour

    // Act
    err := flight.Delay(delay, "Weather")

    // Assert
    assert.Error(t, err)
    assert.Equal(t, entity.ErrCannotDelayDepartedFlight, err)
}
```

### Integration Tests (20%)

Test component interactions with real dependencies.

**What to test:**
- Repository implementations with real database
- External API integrations
- Message queue publishers/consumers

**Example:**
```go
// +build integration

func TestPostgresFlightRepository_Save_Integration(t *testing.T) {
    // Arrange
    db := setupTestDatabase(t)
    defer db.Close()

    repo := postgres.NewFlightRepository(db)
    flight := fixtures.NewFlightBuilder().Build()

    // Act
    err := repo.Save(context.Background(), flight)

    // Assert
    assert.NoError(t, err)

    // Verify in database
    saved, err := repo.FindByID(context.Background(), flight.ID())
    assert.NoError(t, err)
    assert.Equal(t, flight.FlightNumber(), saved.FlightNumber())
}

func setupTestDatabase(t *testing.T) *sql.DB {
    // Use testcontainers for real database
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "postgres:15-alpine",
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "test",
            },
            ExposedPorts: []string{"5432/tcp"},
        },
        Started: true,
    })
    require.NoError(t, err)

    // Get connection details and connect
    // ...

    return db
}
```

### E2E Tests (10%)

Test complete user flows through the system.

**What to test:**
- Critical user journeys
- Happy path scenarios
- Key business workflows

**Example:**
```go
// +build e2e

func TestBookFlightE2E(t *testing.T) {
    // Setup: Start services
    services := startTestServices(t)
    defer services.Cleanup()

    client := http.Client{}
    baseURL := services.APIGatewayURL()

    // Step 1: Search for flights
    searchResp := searchFlights(t, client, baseURL, "JFK", "LAX", tomorrow())
    assert.True(t, len(searchResp.Flights) > 0)

    flightID := searchResp.Flights[0].ID

    // Step 2: Create booking
    bookingResp := createBooking(t, client, baseURL, flightID, passengerData())
    assert.NotEmpty(t, bookingResp.BookingID)
    assert.Equal(t, "PENDING", bookingResp.Status)

    // Step 3: Make payment
    paymentResp := makePayment(t, client, baseURL, bookingResp.BookingID, paymentData())
    assert.Equal(t, "COMPLETED", paymentResp.Status)

    // Step 4: Verify booking confirmed
    booking := getBooking(t, client, baseURL, bookingResp.BookingID)
    assert.Equal(t, "CONFIRMED", booking.Status)

    // Step 5: Check-in
    checkinResp := checkIn(t, client, baseURL, booking.PNR)
    assert.NotEmpty(t, checkinResp.BoardingPass)
}
```

---

## Test Organization

### Directory Structure

```
internal/flight/
├── domain/
│   ├── entity/
│   │   ├── flight.go
│   │   └── flight_test.go         # Unit tests
│   ├── value/
│   │   ├── flight_number.go
│   │   └── flight_number_test.go  # Unit tests
│   └── service/
│       ├── pricing.go
│       └── pricing_test.go        # Unit tests
├── usecase/
│   ├── create_flight.go
│   └── create_flight_test.go      # Unit tests with mocks
├── adapter/
│   ├── http/
│   │   ├── handler.go
│   │   └── handler_test.go        # Unit tests
│   └── repository/
│       ├── postgres_flight_repo.go
│       └── postgres_flight_repo_test.go  # Integration tests
test/
├── fixtures/                       # Test data builders
│   ├── flight_builder.go
│   └── booking_builder.go
├── mocks/                          # Mock implementations
│   ├── flight_repository_mock.go
│   └── payment_gateway_mock.go
├── integration/                    # Integration tests
│   ├── flight_repository_test.go
│   └── booking_flow_test.go
└── e2e/                           # End-to-end tests
    ├── booking_journey_test.go
    └── check_in_journey_test.go
```

---

## TDD Best Practices

### 1. Test Behavior, Not Implementation

```go
// ❌ BAD: Testing implementation details
func TestFlight_InternalState(t *testing.T) {
    flight := &Flight{status: "scheduled"} // Accessing private field
    assert.Equal(t, "scheduled", flight.status)
}

// ✅ GOOD: Testing behavior
func TestFlight_IsScheduled(t *testing.T) {
    flight := fixtures.NewFlightBuilder().Build()
    assert.True(t, flight.IsScheduled())
}
```

### 2. One Assertion Per Test (ideally)

```go
// ❌ BAD: Multiple concerns
func TestCreateFlight_Everything(t *testing.T) {
    flight := createFlight()
    assert.Equal(t, "AA100", flight.FlightNumber())
    assert.Equal(t, "JFK", flight.DepartureCode())
    assert.Equal(t, "scheduled", flight.Status())
    // What failed if test fails?
}

// ✅ GOOD: Focused tests
func TestCreateFlight_SetsFlightNumber(t *testing.T) {
    flight := createFlight("AA100")
    assert.Equal(t, "AA100", flight.FlightNumber())
}

func TestCreateFlight_InitialStatusIsScheduled(t *testing.T) {
    flight := createFlight("AA100")
    assert.Equal(t, "scheduled", flight.Status())
}
```

### 3. Test Names Should Describe Behavior

```go
// ❌ BAD: Unclear what's being tested
func TestFlight1(t *testing.T) {}
func TestFlight2(t *testing.T) {}

// ✅ GOOD: Clear intent
func TestFlight_Delay_UpdatesDepartureTime(t *testing.T) {}
func TestFlight_Delay_ChangesStatusToDelayed(t *testing.T) {}
func TestFlight_Delay_FailsWhenAlreadyDeparted(t *testing.T) {}
```

### 4. Arrange-Act-Assert Pattern

```go
func TestBooking_AddPassenger_Success(t *testing.T) {
    // Arrange: Set up test data
    booking := fixtures.NewBookingBuilder().Build()
    passenger := fixtures.NewPassengerBuilder().Build()

    // Act: Execute the behavior
    err := booking.AddPassenger(passenger)

    // Assert: Verify the outcome
    assert.NoError(t, err)
    assert.Equal(t, 1, booking.PassengerCount())
}
```

### 5. Don't Test Third-Party Code

```go
// ❌ BAD: Testing gin framework
func TestGinReturnsJSON(t *testing.T) {
    router := gin.Default()
    // Testing gin's JSON functionality
}

// ✅ GOOD: Test your handler logic
func TestFlightHandler_CreateFlight_ReturnsCreatedStatus(t *testing.T) {
    handler := NewFlightHandler(mockUseCase)
    req := httptest.NewRequest("POST", "/flights", body)
    w := httptest.NewRecorder()

    handler.CreateFlight(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### 6. Tests Should Be Fast

```go
// ❌ BAD: Slow tests
func TestSlow(t *testing.T) {
    time.Sleep(5 * time.Second) // Don't do this
    // ...
}

// ✅ GOOD: Fast tests
func TestFast(t *testing.T) {
    // No sleep, no real database, no network calls
    result := calculateSomething()
    assert.Equal(t, expected, result)
}
```

### 7. Tests Should Be Independent

```go
// ❌ BAD: Tests depend on each other
var sharedFlight *Flight

func TestA(t *testing.T) {
    sharedFlight = createFlight()
    sharedFlight.Delay(2 * time.Hour)
}

func TestB(t *testing.T) {
    // Depends on TestA running first!
    assert.True(t, sharedFlight.IsDelayed())
}

// ✅ GOOD: Independent tests
func TestA(t *testing.T) {
    flight := createFlight()
    flight.Delay(2 * time.Hour)
    assert.True(t, flight.IsDelayed())
}

func TestB(t *testing.T) {
    flight := createFlight()
    flight.Delay(2 * time.Hour)
    assert.True(t, flight.IsDelayed())
}
```

---

## Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/flight/domain/entity/...

# Run tests with coverage
make test-coverage

# Run only unit tests
go test -short ./...

# Run integration tests
go test -tags=integration ./test/integration/...

# Run E2E tests
go test -tags=e2e ./test/e2e/...

# Run tests matching pattern
go test -run TestFlight_Delay ./...

# Verbose output
go test -v ./...

# Run tests in parallel
go test -parallel 4 ./...
```

---

## Summary

### TDD Benefits
- ✅ Better design (forces you to think about interfaces)
- ✅ Living documentation (tests show how to use code)
- ✅ Regression protection
- ✅ Confidence to refactor
- ✅ Fewer bugs

### TDD Workflow
1. **RED**: Write failing test
2. **GREEN**: Make it pass (quickly)
3. **REFACTOR**: Improve code quality
4. **REPEAT**: Iterate

### Key Principles
- Write tests first
- One test at a time
- Keep tests simple and focused
- Make tests fast and independent
- Test behavior, not implementation

By following TDD, we ensure the Airport Services platform is reliable, maintainable, and well-designed from the start.

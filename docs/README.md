# Airport Services - Architecture Documentation

## Overview

This directory contains comprehensive architectural guidelines and best practices for the Airport Services microservices platform. All services in this repository follow these principles to ensure consistency, maintainability, and quality.

## Documentation Index

### 📐 Core Architectural Patterns

1. **[Clean Architecture Guide](./CLEAN_ARCHITECTURE.md)**
   - Layer structure and dependencies
   - Entities, Use Cases, Interface Adapters, Frameworks
   - Dependency Inversion and interface design
   - Complete code examples with Flight Service

2. **[Domain-Driven Design Guide](./DDD_GUIDE.md)**
   - Ubiquitous Language
   - Bounded Contexts
   - Entities vs Value Objects
   - Aggregates and Aggregate Roots
   - Domain Events and Services
   - Repositories and Factories
   - Strategic design patterns

3. **[SOLID Principles Guide](./SOLID_PRINCIPLES.md)**
   - Single Responsibility Principle (SRP)
   - Open/Closed Principle (OCP)
   - Liskov Substitution Principle (LSP)
   - Interface Segregation Principle (ISP)
   - Dependency Inversion Principle (DIP)
   - Practical Go examples from Airport domain

4. **[Test-Driven Development Guide](./TDD_GUIDE.md)**
   - Red-Green-Refactor cycle
   - Complete TDD examples
   - Test patterns (Table-driven, Fixtures, Mocks)
   - Testing pyramid (Unit, Integration, E2E)
   - Best practices and anti-patterns

## Quick Reference

### Project Structure (Clean Architecture)

```
internal/{service}/
├── domain/                    # Layer 1: Entities & Business Logic
│   ├── entity/               # Domain entities (Flight, Booking)
│   ├── value/                # Value objects (FlightNumber, Money)
│   ├── repository/           # Repository interfaces (ports)
│   ├── event/                # Domain events
│   └── service/              # Domain services
│
├── usecase/                  # Layer 2: Application Logic
│   ├── create_flight.go      # Use case implementation
│   └── create_flight_test.go # Use case tests
│
├── adapter/                  # Layer 3: Interface Adapters
│   ├── http/                 # HTTP controllers/handlers
│   ├── grpc/                 # gRPC controllers
│   ├── repository/           # Repository implementations
│   ├── event/                # Event publishers/subscribers
│   └── presenter/            # Response formatters
│
└── infrastructure/           # Layer 4: External Concerns (optional)
    └── cache/                # Caching implementations
```

### Dependency Flow

```
Frameworks & Drivers
    ↓ (depends on)
Interface Adapters
    ↓ (depends on)
Use Cases
    ↓ (depends on)
Entities (Domain)
```

**Key Rule**: Dependencies always point inward. Outer layers depend on inner layers, never the reverse.

---

## Development Workflow

### 1. Planning a New Feature

Before writing code:
1. ✅ Identify the **Bounded Context** (which service?)
2. ✅ Define **Ubiquitous Language** (domain terms)
3. ✅ Identify **Entities** vs **Value Objects**
4. ✅ Define **Aggregates** and their boundaries
5. ✅ Design **Use Cases** (application logic)
6. ✅ Define **Repository Interfaces** (ports)
7. ✅ Plan **Domain Events** for communication

### 2. Implementing with TDD

Follow the Red-Green-Refactor cycle:

```bash
# 1. RED: Write failing test
# internal/flight/domain/entity/flight_test.go
func TestFlight_Delay_UpdatesTime(t *testing.T) {
    flight := NewFlight(...)
    err := flight.Delay(2 * time.Hour, "Weather")
    assert.NoError(t, err)
    // Test fails - Delay method doesn't exist
}

# 2. GREEN: Write minimal code to pass
# internal/flight/domain/entity/flight.go
func (f *Flight) Delay(duration time.Duration, reason string) error {
    f.departureTime = f.departureTime.Add(duration)
    return nil
}
# Test passes!

# 3. REFACTOR: Improve code quality
func (f *Flight) Delay(duration time.Duration, reason string) error {
    if !f.CanBeDelayed() {
        return ErrCannotDelay
    }
    f.departureTime = f.departureTime.Add(duration)
    f.status = FlightStatusDelayed
    return nil
}
# All tests still pass!
```

### 3. Layer-by-Layer Implementation

Start from the inside and work outward:

**Layer 1: Domain (Entities, Value Objects)**
```go
// 1. Define entity with business logic
type Flight struct {
    id            FlightID
    flightNumber  FlightNumber  // Value object
    status        FlightStatus
}

func (f *Flight) Delay(duration time.Duration) error {
    // Business rules here
}
```

**Layer 2: Use Cases**
```go
// 2. Define use case
type DelayFlightUseCase struct {
    flightRepo repository.FlightRepository  // Interface
}

func (uc *DelayFlightUseCase) Execute(ctx context.Context, input Input) error {
    flight, _ := uc.flightRepo.FindByID(ctx, input.FlightID)
    flight.Delay(input.Duration)
    return uc.flightRepo.Save(ctx, flight)
}
```

**Layer 3: Adapters**
```go
// 3. HTTP Handler
func (h *FlightHandler) DelayFlight(c *gin.Context) {
    var req DelayFlightRequest
    c.BindJSON(&req)

    input := toUseCaseInput(req)
    err := h.delayFlightUC.Execute(c.Request.Context(), input)

    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"message": "Flight delayed"})
}

// 4. Repository Implementation
type PostgresFlightRepository struct {
    db *sql.DB
}

func (r *PostgresFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    // Database logic
}
```

**Layer 4: Main (Wiring)**
```go
// 5. Dependency injection
func main() {
    db := setupDatabase()
    flightRepo := postgres.NewFlightRepository(db)
    delayFlightUC := usecase.NewDelayFlightUseCase(flightRepo)
    handler := http.NewFlightHandler(delayFlightUC)

    router := gin.Default()
    router.PUT("/flights/:id/delay", handler.DelayFlight)
    router.Run(":8081")
}
```

---

## Design Checklist

Use this checklist when designing a new service or feature:

### ✅ Clean Architecture
- [ ] Business logic in domain layer (no framework dependencies)
- [ ] Use cases orchestrate entities
- [ ] Handlers/controllers are thin (just adapters)
- [ ] Dependencies point inward
- [ ] Interfaces defined in domain layer
- [ ] Implementations in adapter/infrastructure layer

### ✅ DDD
- [ ] Service belongs to a clear Bounded Context
- [ ] Ubiquitous Language used consistently
- [ ] Rich domain models (behavior, not just data)
- [ ] Aggregates enforce invariants
- [ ] Domain events for significant occurrences
- [ ] Repository per Aggregate Root

### ✅ SOLID
- [ ] Each struct has single responsibility
- [ ] New behavior added via extension (interfaces), not modification
- [ ] Implementations honor interface contracts
- [ ] Small, focused interfaces
- [ ] Dependencies injected via constructors (DI)

### ✅ TDD
- [ ] Tests written before production code
- [ ] Unit tests for domain logic (70%)
- [ ] Integration tests for adapters (20%)
- [ ] E2E tests for critical flows (10%)
- [ ] All tests pass before committing
- [ ] Code coverage >80%

---

## Code Review Checklist

When reviewing code, check for:

### Architecture
- [ ] Follows Clean Architecture layers
- [ ] No domain code depends on frameworks
- [ ] Use cases don't know about HTTP/database details
- [ ] Proper dependency direction

### DDD
- [ ] Domain terms match Ubiquitous Language
- [ ] Entities have behavior, not just getters/setters
- [ ] Business rules in domain, not in handlers
- [ ] Aggregates protect invariants
- [ ] Events for significant domain occurrences

### SOLID
- [ ] Single Responsibility: each file/struct one purpose
- [ ] Interfaces for extension points
- [ ] No LSP violations (subtypes work correctly)
- [ ] Small interfaces, not fat ones
- [ ] Dependency Injection used

### Testing
- [ ] Tests written (TDD followed)
- [ ] Tests are fast and independent
- [ ] Mocks used for external dependencies
- [ ] Test names describe behavior
- [ ] Arrange-Act-Assert pattern

### General
- [ ] Code is readable and well-documented
- [ ] No premature optimization
- [ ] Error handling consistent
- [ ] Logging appropriate
- [ ] No secrets in code

---

## Common Patterns

### Use Case Pattern

```go
type CreateBookingUseCase struct {
    bookingRepo  repository.BookingRepository
    paymentGW    PaymentGateway
    eventBus     EventBus
}

func (uc *CreateBookingUseCase) Execute(ctx context.Context, input Input) (*Output, error) {
    // 1. Validate input
    // 2. Create domain entity
    // 3. Call domain methods
    // 4. Persist via repository
    // 5. Publish domain events
    // 6. Return output DTO
}
```

### Repository Pattern

```go
// Interface in domain layer
type FlightRepository interface {
    Save(ctx context.Context, flight *entity.Flight) error
    FindByID(ctx context.Context, id string) (*entity.Flight, error)
}

// Implementation in adapter layer
type PostgresFlightRepository struct {
    db *sql.DB
}

func (r *PostgresFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    // Convert entity to database model
    // Execute SQL
}
```

### Value Object Pattern

```go
type Money struct {
    amount   decimal.Decimal
    currency Currency
}

func NewMoney(amount decimal.Decimal, currency Currency) Money {
    return Money{amount: amount, currency: currency}
}

// Immutable - returns new instance
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return NewMoney(m.amount.Add(other.amount), m.currency), nil
}
```

### Aggregate Pattern

```go
type Booking struct {
    id         BookingID  // Aggregate Root ID
    passengers []Passenger  // Internal entities
    status     BookingStatus
}

// Only root is accessed externally
func (b *Booking) AddPassenger(info PassengerInfo) error {
    // Enforce invariants
    if len(b.passengers) >= MaxPassengers {
        return ErrTooManyPassengers
    }
    passenger := NewPassenger(info)
    b.passengers = append(b.passengers, passenger)
    return nil
}
```

---

## Learning Resources

### Books
- **Clean Architecture** by Robert C. Martin
- **Domain-Driven Design** by Eric Evans
- **Implementing Domain-Driven Design** by Vaughn Vernon
- **Test Driven Development: By Example** by Kent Beck

### Online Resources
- [Clean Architecture Blog Post](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [DDD Reference](https://www.domainlanguage.com/ddd/reference/)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Go Testing Best Practices](https://golang.org/doc/tutorial/add-a-test)

### Internal Resources
- Architecture Decision Records (ADRs) in `docs/adr/`
- Service templates in `templates/`
- Code examples throughout `docs/` guides

---

## Questions?

- Review the specific guides for detailed examples
- Check existing services for reference implementations
- Ask in team discussions or code reviews
- Refer to ADRs for architecture decisions

---

**Remember**: These are not just rules to follow, but principles that help us write better, more maintainable code. When in doubt, favor simplicity and clarity over cleverness.

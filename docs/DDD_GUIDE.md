# Domain-Driven Design (DDD) Implementation Guide

## Overview

Domain-Driven Design is our strategic approach to building the Airport Services platform. DDD helps us manage complexity by focusing on the core domain and domain logic, using a common language (Ubiquitous Language), and organizing code around business capabilities.

## Core DDD Concepts

### 1. Ubiquitous Language

A common, rigorous language between developers and domain experts, used consistently in code, documentation, and conversation.

**Example - Flight Domain:**
```
✅ Use Domain Terms:
- "Flight" not "Trip"
- "Boarding" not "Getting on plane"
- "PNR" (Passenger Name Record) not "Booking ID"
- "Check-in" not "Registration"
- "Gate" not "Door"
- "Delay" not "Late"

✅ Code Reflects Language:
type Flight struct {
    flightNumber  string    // Not "id" or "code"
    status        FlightStatus
    boardingTime  time.Time
    gateNumber    string
}

func (f *Flight) DelayFlight(duration time.Duration)  // Not "postpone" or "defer"
func (f *Flight) BoardPassenger(passenger Passenger)  // Exact domain term
```

### 2. Bounded Context

A boundary within which a particular domain model is defined and applicable. Different contexts may have different meanings for the same term.

**Airport Services Bounded Contexts:**

```
┌─────────────────────────────────────────────────────────┐
│                  Airport Services                        │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Flight     │  │  Passenger   │  │   Baggage    │  │
│  │   Context    │  │   Context    │  │   Context    │  │
│  │              │  │              │  │              │  │
│  │ - Flight     │  │ - Passenger  │  │ - Bag        │  │
│  │ - Aircraft   │  │ - Booking    │  │ - Tag        │  │
│  │ - Crew       │  │ - Check-in   │  │ - Routing    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│         │                  │                  │         │
│         └──────────────────┴──────────────────┘         │
│                            │                            │
│                  ┌──────────────────┐                   │
│                  │   Payment        │                   │
│                  │   Context        │                   │
│                  │                  │                   │
│                  │ - Transaction    │                   │
│                  │ - Refund         │                   │
│                  └──────────────────┘                   │
└─────────────────────────────────────────────────────────┘
```

**Context Map Example:**
```go
// Flight Context - "Passenger" means someone on a flight
type Passenger struct {
    id           string
    seatNumber   string
    boardingZone int
}

// Passenger Context - "Passenger" means customer with full profile
type Passenger struct {
    id              string
    personalInfo    PersonalInfo
    travelDocuments []Document
    preferences     Preferences
}

// Baggage Context - "Passenger" is just a reference
type PassengerID string  // Just an ID, not full entity
```

---

## DDD Building Blocks

### 1. Entities

Objects with a distinct identity that persists over time.

**Characteristics:**
- Has a unique identifier
- Identity is important, not just attributes
- Mutable
- Lifecycle tracked

**Example:**
```go
// internal/flight/domain/entity/flight.go
package entity

// Flight is an entity - identity matters
type Flight struct {
    id            FlightID      // Unique identity
    flightNumber  string
    departureTime time.Time
    status        FlightStatus
    version       int           // For optimistic locking
}

// NewFlight creates a new flight (factory method)
func NewFlight(flightNumber string, departure, arrival Airport,
               departureTime, arrivalTime time.Time) (*Flight, error) {

    // Validate business rules
    if err := validateFlightNumber(flightNumber); err != nil {
        return nil, err
    }

    if err := validateTimes(departureTime, arrivalTime); err != nil {
        return nil, err
    }

    return &Flight{
        id:            NewFlightID(),
        flightNumber:  flightNumber,
        departure:     departure,
        arrival:       arrival,
        departureTime: departureTime,
        arrivalTime:   arrivalTime,
        status:        FlightStatusScheduled,
        version:       1,
    }, nil
}

// Identity method
func (f *Flight) ID() FlightID {
    return f.id
}

// Equals compares identity, not attributes
func (f *Flight) Equals(other *Flight) bool {
    return f.id == other.id
}

// Business methods
func (f *Flight) Delay(duration time.Duration, reason string) error {
    if !f.CanBeDelayed() {
        return ErrCannotDelayFlight
    }

    f.departureTime = f.departureTime.Add(duration)
    f.arrivalTime = f.arrivalTime.Add(duration)
    f.status = FlightStatusDelayed
    f.version++

    return nil
}
```

### 2. Value Objects

Objects that describe characteristics but have no identity.

**Characteristics:**
- No unique identifier
- Immutable
- Defined by attributes
- Can be shared

**Examples:**
```go
// internal/flight/domain/value/flight_number.go
package value

// FlightNumber is a value object
type FlightNumber struct {
    airlineCode string  // e.g., "AA"
    number      string  // e.g., "100"
}

// NewFlightNumber creates and validates a flight number
func NewFlightNumber(code string) (FlightNumber, error) {
    parts := parseFlightNumber(code)
    if !isValidAirlineCode(parts.airline) {
        return FlightNumber{}, ErrInvalidAirlineCode
    }
    if !isValidNumber(parts.number) {
        return FlightNumber{}, ErrInvalidFlightNumber
    }

    return FlightNumber{
        airlineCode: parts.airline,
        number:      parts.number,
    }, nil
}

// Value objects are immutable - no setters
func (fn FlightNumber) Code() string {
    return fn.airlineCode + fn.number
}

func (fn FlightNumber) AirlineCode() string {
    return fn.airlineCode
}

// Equals compares values, not identity
func (fn FlightNumber) Equals(other FlightNumber) bool {
    return fn.airlineCode == other.airlineCode &&
           fn.number == other.number
}

// internal/flight/domain/value/money.go
package value

// Money is a value object
type Money struct {
    amount   decimal.Decimal
    currency Currency
}

func NewMoney(amount decimal.Decimal, currency Currency) Money {
    return Money{amount: amount, currency: currency}
}

// Immutable operations return new instances
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{
        amount:   m.amount.Add(other.amount),
        currency: m.currency,
    }, nil
}

func (m Money) MultiplyBy(factor decimal.Decimal) Money {
    return Money{
        amount:   m.amount.Mul(factor),
        currency: m.currency,
    }
}

// internal/passenger/domain/value/email.go
package value

type Email struct {
    address string
}

func NewEmail(address string) (Email, error) {
    if !isValidEmail(address) {
        return Email{}, ErrInvalidEmail
    }
    return Email{address: strings.ToLower(address)}, nil
}

func (e Email) String() string {
    return e.address
}

// internal/passenger/domain/value/passport.go
package value

type Passport struct {
    number      string
    country     CountryCode
    issueDate   time.Time
    expiryDate  time.Time
}

func NewPassport(number string, country CountryCode, issue, expiry time.Time) (Passport, error) {
    if expiry.Before(issue) {
        return Passport{}, ErrInvalidPassportDates
    }
    return Passport{
        number:     number,
        country:    country,
        issueDate:  issue,
        expiryDate: expiry,
    }, nil
}

func (p Passport) IsValid(asOf time.Time) bool {
    return asOf.Before(p.expiryDate) && asOf.After(p.issueDate)
}
```

### 3. Aggregates

A cluster of entities and value objects with a root entity (Aggregate Root).

**Rules:**
- One root entity controls access
- External objects can only reference the root
- Root enforces invariants
- Transactional consistency boundary

**Example:**
```go
// internal/booking/domain/aggregate/booking.go
package aggregate

// Booking is an Aggregate Root
type Booking struct {
    // Aggregate Root Identity
    id  BookingID
    pnr PNR  // Passenger Name Record

    // Entities within aggregate
    passengers []Passenger      // Internal entities
    segments   []FlightSegment  // Internal entities

    // Value Objects
    bookingDate   time.Time
    totalPrice    Money
    status        BookingStatus
    contactInfo   ContactInfo

    // Aggregate metadata
    version int
}

// NewBooking is the factory method (Aggregate Root constructor)
func NewBooking(passengers []PassengerInfo, flights []FlightInfo) (*Booking, error) {
    // Validate aggregate invariants
    if len(passengers) == 0 {
        return nil, ErrNoPassengers
    }
    if len(flights) == 0 {
        return nil, ErrNoFlights
    }

    // Create aggregate
    booking := &Booking{
        id:          NewBookingID(),
        pnr:         GeneratePNR(),
        passengers:  make([]Passenger, 0),
        segments:    make([]FlightSegment, 0),
        bookingDate: time.Now(),
        status:      BookingStatusPending,
        version:     1,
    }

    // Add passengers through the root (maintains invariants)
    for _, info := range passengers {
        if err := booking.AddPassenger(info); err != nil {
            return nil, err
        }
    }

    // Add segments through the root
    for _, flight := range flights {
        if err := booking.AddSegment(flight); err != nil {
            return nil, err
        }
    }

    return booking, nil
}

// AddPassenger adds a passenger (only through root)
func (b *Booking) AddPassenger(info PassengerInfo) error {
    // Enforce aggregate invariants
    if b.status != BookingStatusPending {
        return ErrCannotModifyConfirmedBooking
    }

    if len(b.passengers) >= MaxPassengersPerBooking {
        return ErrTooManyPassengers
    }

    passenger := NewPassenger(info)
    b.passengers = append(b.passengers, passenger)
    b.version++

    return nil
}

// Confirm confirms the booking (aggregate-level operation)
func (b *Booking) Confirm(payment Payment) error {
    // Validate aggregate state
    if b.status != BookingStatusPending {
        return ErrInvalidBookingStatus
    }

    if len(b.passengers) == 0 {
        return ErrNoPassengers
    }

    // Verify payment amount matches total
    if !payment.Amount().Equals(b.totalPrice) {
        return ErrPaymentMismatch
    }

    // Change aggregate state
    b.status = BookingStatusConfirmed
    b.version++

    return nil
}

// Cancel cancels the booking
func (b *Booking) Cancel(reason string) error {
    if !b.CanBeCancelled() {
        return ErrCannotCancelBooking
    }

    b.status = BookingStatusCancelled
    b.version++

    return nil
}

// CanBeCancelled checks if cancellation is allowed (business rule)
func (b *Booking) CanBeCancelled() bool {
    return b.status == BookingStatusPending ||
           b.status == BookingStatusConfirmed
}

// GetPassenger gets passenger by ID (controlled access)
func (b *Booking) GetPassenger(id PassengerID) (*Passenger, error) {
    for i := range b.passengers {
        if b.passengers[i].ID() == id {
            return &b.passengers[i], nil
        }
    }
    return nil, ErrPassengerNotFound
}

// Passengers returns a copy (protect encapsulation)
func (b *Booking) Passengers() []Passenger {
    return append([]Passenger{}, b.passengers...)
}

// Aggregate Root getters
func (b *Booking) ID() BookingID { return b.id }
func (b *Booking) PNR() PNR { return b.pnr }
func (b *Booking) Status() BookingStatus { return b.status }
func (b *Booking) TotalPrice() Money { return b.totalPrice }
```

**Key Points:**
- Only `Booking` (root) is accessed from outside
- `Passenger` and `FlightSegment` are managed through `Booking`
- All invariants are enforced by the root
- Changes to aggregate go through root methods

### 4. Domain Services

Operations that don't naturally belong to an entity or value object.

**When to use:**
- Operation involves multiple aggregates
- Operation is stateless
- Significant business logic that doesn't fit in an entity

**Examples:**
```go
// internal/flight/domain/service/flight_pricing_service.go
package service

// FlightPricingService calculates flight prices (domain service)
type FlightPricingService struct {
    pricingRules PricingRuleRepository
}

func NewFlightPricingService(rules PricingRuleRepository) *FlightPricingService {
    return &FlightPricingService{pricingRules: rules}
}

// CalculatePrice is a domain service operation
func (s *FlightPricingService) CalculatePrice(
    flight *Flight,
    passenger *Passenger,
    bookingDate time.Time,
) (Money, error) {

    basePrice := flight.BasePrice()

    // Apply date-based pricing
    daysUntilDeparture := flight.DepartureTime().Sub(bookingDate).Hours() / 24
    if daysUntilDeparture < 7 {
        basePrice = basePrice.MultiplyBy(decimal.NewFromFloat(1.5))
    }

    // Apply passenger-specific discounts
    if passenger.IsSenior() {
        basePrice = basePrice.MultiplyBy(decimal.NewFromFloat(0.9))
    }

    // Apply route-specific rules
    rules := s.pricingRules.FindByRoute(flight.Route())
    for _, rule := range rules {
        basePrice = rule.Apply(basePrice)
    }

    return basePrice, nil
}

// internal/passenger/domain/service/duplicate_detection_service.go
package service

// DuplicateDetectionService detects duplicate passengers
type DuplicateDetectionService struct {
    passengerRepo PassengerRepository
}

func NewDuplicateDetectionService(repo PassengerRepository) *DuplicateDetectionService {
    return &DuplicateDetectionService{passengerRepo: repo}
}

// IsDuplicate checks if passenger already exists
func (s *DuplicateDetectionService) IsDuplicate(
    ctx context.Context,
    candidate *Passenger,
) (bool, error) {

    // Check by passport
    if existing, _ := s.passengerRepo.FindByPassport(ctx, candidate.Passport()); existing != nil {
        return true, nil
    }

    // Check by email + name
    if existing, _ := s.passengerRepo.FindByEmailAndName(
        ctx,
        candidate.Email(),
        candidate.FullName(),
    ); existing != nil {
        return true, nil
    }

    return false, nil
}
```

### 5. Domain Events

Something that happened in the domain that domain experts care about.

**Characteristics:**
- Named in past tense
- Immutable
- Contains event-specific data
- Timestamp

**Examples:**
```go
// internal/flight/domain/event/flight_delayed.go
package event

// FlightDelayed is a domain event
type FlightDelayed struct {
    eventID       string
    occurredAt    time.Time
    flightID      string
    flightNumber  string
    originalTime  time.Time
    newTime       time.Time
    delayDuration time.Duration
    reason        string
}

func NewFlightDelayed(flight *Flight, newTime time.Time, reason string) FlightDelayed {
    return FlightDelayed{
        eventID:       uuid.New().String(),
        occurredAt:    time.Now(),
        flightID:      flight.ID().String(),
        flightNumber:  flight.FlightNumber(),
        originalTime:  flight.DepartureTime(),
        newTime:       newTime,
        delayDuration: newTime.Sub(flight.DepartureTime()),
        reason:        reason,
    }
}

// Event interface implementation
func (e FlightDelayed) EventID() string { return e.eventID }
func (e FlightDelayed) OccurredAt() time.Time { return e.occurredAt }
func (e FlightDelayed) EventType() string { return "flight.delayed.v1" }

// internal/booking/domain/event/booking_confirmed.go
package event

type BookingConfirmed struct {
    eventID     string
    occurredAt  time.Time
    bookingID   string
    pnr         string
    passengerIDs []string
    totalAmount Money
}

func NewBookingConfirmed(booking *Booking) BookingConfirmed {
    return BookingConfirmed{
        eventID:     uuid.New().String(),
        occurredAt:  time.Now(),
        bookingID:   booking.ID().String(),
        pnr:         booking.PNR().String(),
        passengerIDs: extractPassengerIDs(booking.Passengers()),
        totalAmount: booking.TotalPrice(),
    }
}
```

**Publishing Events:**
```go
// In aggregate
func (f *Flight) Delay(duration time.Duration, reason string) error {
    if err := f.validate(); err != nil {
        return err
    }

    f.departureTime = f.departureTime.Add(duration)
    f.status = FlightStatusDelayed

    // Record domain event (not published yet)
    f.recordEvent(NewFlightDelayed(f, f.departureTime, reason))

    return nil
}

// In use case (publish events)
func (uc *DelayFlightUseCase) Execute(ctx context.Context, input Input) error {
    flight, err := uc.flightRepo.FindByID(ctx, input.FlightID)
    if err != nil {
        return err
    }

    if err := flight.Delay(input.Duration, input.Reason); err != nil {
        return err
    }

    if err := uc.flightRepo.Save(ctx, flight); err != nil {
        return err
    }

    // Publish domain events
    for _, event := range flight.Events() {
        uc.eventBus.Publish(ctx, event)
    }

    return nil
}
```

### 6. Repositories

Abstraction for accessing aggregates (collections of domain objects).

**Rules:**
- Repository per Aggregate Root only
- Repository interface in domain layer
- Implementation in infrastructure layer
- Collection-like interface

**Example:**
```go
// internal/booking/domain/repository/booking_repository.go
package repository

// BookingRepository manages Booking aggregates
type BookingRepository interface {
    // Save persists a booking aggregate
    Save(ctx context.Context, booking *aggregate.Booking) error

    // FindByID retrieves by aggregate root ID
    FindByID(ctx context.Context, id BookingID) (*aggregate.Booking, error)

    // FindByPNR retrieves by PNR (business identifier)
    FindByPNR(ctx context.Context, pnr PNR) (*aggregate.Booking, error)

    // Delete removes a booking
    Delete(ctx context.Context, id BookingID) error

    // NextID generates next identity
    NextID() BookingID
}
```

### 7. Factories

Complex object creation logic.

**Examples:**
```go
// internal/booking/domain/factory/booking_factory.go
package factory

// BookingFactory creates complex Booking aggregates
type BookingFactory struct {
    pricingService    *service.FlightPricingService
    passengerValidator *service.PassengerValidator
}

func NewBookingFactory(
    pricingService *service.FlightPricingService,
    validator *service.PassengerValidator,
) *BookingFactory {
    return &BookingFactory{
        pricingService:    pricingService,
        passengerValidator: validator,
    }
}

// CreateRoundTripBooking creates a round-trip booking
func (f *BookingFactory) CreateRoundTripBooking(
    passengers []PassengerInfo,
    outbound FlightInfo,
    inbound FlightInfo,
) (*aggregate.Booking, error) {

    // Validate passengers
    for _, passenger := range passengers {
        if err := f.passengerValidator.Validate(passenger); err != nil {
            return nil, err
        }
    }

    // Create booking
    booking := aggregate.NewBooking(passengers, []FlightInfo{outbound, inbound})

    // Calculate pricing
    for _, passenger := range booking.Passengers() {
        price, _ := f.pricingService.CalculatePrice(outbound, passenger)
        booking.AddPrice(price)
    }

    return booking, nil
}
```

---

## Strategic Design Patterns

### 1. Context Mapping

Define relationships between bounded contexts:

```go
// Flight Context → Passenger Context (Customer-Supplier)
// Flight context needs passenger information

// Anti-Corruption Layer in Flight Context
type PassengerInfo struct {
    id       string
    name     string
    seatPref SeatPreference
}

// PassengerAdapter translates between contexts
type PassengerAdapter struct {
    passengerClient PassengerServiceClient
}

func (a *PassengerAdapter) GetPassengerInfo(passengerID string) (*PassengerInfo, error) {
    // Call Passenger Context API
    passengerDTO, err := a.passengerClient.GetPassenger(passengerID)
    if err != nil {
        return nil, err
    }

    // Translate to Flight Context model (Anti-Corruption Layer)
    return &PassengerInfo{
        id:       passengerDTO.ID,
        name:     passengerDTO.FirstName + " " + passengerDTO.LastName,
        seatPref: translateSeatPreference(passengerDTO.Preferences),
    }, nil
}
```

### 2. Shared Kernel

Shared domain model between contexts (use sparingly):

```go
// pkg/shared/domain/types/airport_code.go
package types

// AirportCode is shared across contexts
type AirportCode string

func (a AirportCode) Validate() error {
    if len(a) != 3 {
        return errors.New("airport code must be 3 characters")
    }
    return nil
}
```

---

## Practical Examples

### Complete Aggregate Example

```go
// internal/baggage/domain/aggregate/baggage.go
package aggregate

// Baggage is an Aggregate Root
type Baggage struct {
    // Identity
    id         BaggageID
    tagNumber  TagNumber

    // Aggregate internals
    routingTags []RoutingTag  // Value objects
    scans       []Scan         // Entities

    // Value objects
    weight      Weight
    dimensions  Dimensions
    owner       PassengerID
    status      BaggageStatus

    // Metadata
    version int
}

func NewBaggage(owner PassengerID, weight Weight, dimensions Dimensions) (*Baggage, error) {
    if err := weight.Validate(); err != nil {
        return nil, err
    }

    return &Baggage{
        id:         NewBaggageID(),
        tagNumber:  GenerateTagNumber(),
        owner:      owner,
        weight:     weight,
        dimensions: dimensions,
        status:     BaggageStatusCheckedIn,
        scans:      make([]Scan, 0),
        routingTags: make([]RoutingTag, 0),
        version:    1,
    }, nil
}

func (b *Baggage) AddRoutingTag(destination AirportCode, flight FlightNumber) error {
    tag := NewRoutingTag(destination, flight)
    b.routingTags = append(b.routingTags, tag)
    b.version++
    return nil
}

func (b *Baggage) RecordScan(location string, scanType ScanType) error {
    scan := NewScan(location, scanType, time.Now())
    b.scans = append(b.scans, scan)
    b.updateStatusFromScan(scanType)
    b.version++
    return nil
}

func (b *Baggage) updateStatusFromScan(scanType ScanType) {
    switch scanType {
    case ScanTypeLoaded:
        b.status = BaggageStatusInTransit
    case ScanTypeUnloaded:
        b.status = BaggageStatusArrived
    case ScanTypeDelivered:
        b.status = BaggageStatusDelivered
    }
}

func (b *Baggage) MarkAsLost() error {
    if b.status == BaggageStatusDelivered {
        return ErrCannotMarkDeliveredAsLost
    }
    b.status = BaggageStatusLost
    b.version++
    return nil
}
```

---

## Anti-Patterns to Avoid

### ❌ Anemic Domain Model
```go
// BAD: No behavior, just data
type Flight struct {
    ID           string
    FlightNumber string
    Status       string
}

// Business logic in service instead of domain
func (s *FlightService) DelayFlight(flightID string, duration time.Duration) {
    flight := s.repo.FindByID(flightID)
    flight.Status = "DELAYED"
    s.repo.Save(flight)
}
```

### ✅ Rich Domain Model
```go
// GOOD: Behavior in domain
type Flight struct {
    id     string
    status FlightStatus
}

func (f *Flight) Delay(duration time.Duration) error {
    if !f.CanBeDelayed() {
        return ErrCannotDelay
    }
    f.status = FlightStatusDelayed
    return nil
}
```

---

## Summary

DDD helps us:
1. **Focus on Core Domain**: Business logic first
2. **Use Ubiquitous Language**: Common terminology
3. **Define Bounded Contexts**: Clear boundaries
4. **Model with Rich Entities**: Behavior, not just data
5. **Protect Invariants**: Aggregates enforce rules
6. **Communicate with Events**: Loosely coupled domains
7. **Isolate Persistence**: Repositories abstract storage

By following DDD, we create software that truly reflects the business domain and remains maintainable as it grows.

# Go Microservices Best Practices - Research Synthesis

## Overview

This document synthesizes research from **40+ authoritative sources** (2024-2025) on Go microservices best practices. These guidelines are mandatory for all Airport Services platform development.

## Table of Contents

1. [Go Microservices Architecture](#go-microservices-architecture)
2. [Project Structure & Organization](#project-structure--organization)
3. [Clean Architecture Implementation](#clean-architecture-implementation)
4. [Domain-Driven Design Patterns](#domain-driven-design-patterns)
5. [SOLID Principles in Go](#solid-principles-in-go)
6. [Test-Driven Development](#test-driven-development)
7. [API Design & Versioning](#api-design--versioning)
8. [Communication Patterns](#communication-patterns)
9. [Data Management](#data-management)
10. [Error Handling](#error-handling)
11. [Concurrency Patterns](#concurrency-patterns)
12. [Observability & Monitoring](#observability--monitoring)
13. [Security Best Practices](#security-best-practices)
14. [Performance Optimization](#performance-optimization)
15. [Deployment & Infrastructure](#deployment--infrastructure)

---

## 1. Go Microservices Architecture

### Key Trends (2025)

**Adoption & Popularity:**
- 48% of Go developers use Gin framework (leading choice)
- Go remains in top 4 languages for microservices (alongside Java, Python, Rust)
- 11% of developers plan to adopt Go in next 12 months

**Why Go Excels for Microservices:**
- ✅ Compiles to native machine code (extremely fast)
- ✅ Low memory footprint
- ✅ Goroutines handle thousands of concurrent requests effortlessly
- ✅ Built-in concurrency model (CSP - Communicating Sequential Processes)

### Core Design Principles

1. **Keep Services Small and Focused**
   - Each service: single responsibility
   - Organize around business domains (DDD)

2. **Database per Service**
   - Each microservice has its own database
   - Ensures loose coupling and independent scalability

3. **Communication Patterns**
   - RESTful APIs still dominant
   - GraphQL and gRPC gaining popularity
   - gRPC is 5-8x faster than REST+JSON

4. **Resilience Patterns**
   - Circuit breakers (go-breaker library)
   - Retries with exponential backoff
   - Graceful degradation

**Source References:** JetBrains Go Ecosystem 2025, Medium (Building Scalable Microservices with Go), InfoQ (Using Golang at The Economist)

---

## 2. Project Structure & Organization

### golang-standards/project-layout

**Key Directories:**

```
/cmd              # Application entry points
/internal         # Private application code
/pkg              # Public libraries (use sparingly)
/api              # API definitions (proto, openapi)
/deployments      # Docker, K8s, Terraform
/scripts          # Build and utility scripts
/test             # Integration, E2E tests
/docs             # Documentation
```

### Best Practices

✅ **DO:**
- Use `/internal` for non-exportable packages
- Place application entry points in `/cmd/{service-name}/main.go`
- Use `/api` for API contracts (protobuf, OpenAPI)

❌ **DON'T:**
- Use `/src` directory (unnecessary in Go)
- Put excessive code in application directory
- Overuse `/pkg` (only for truly reusable code)

### Ben Johnson's Approach

- Organize by bounded context/domain
- Put domain-specific code together
- Keep implementation details separate from domain logic

**Source References:** golang-standards/project-layout (GitHub), AppliedGo, Go.dev official docs, Alex Edwards

---

## 3. Clean Architecture Implementation

### Layer Structure

```
Domain Layer (Entities)
    ↑
Use Cases (Application Logic)
    ↑
Interface Adapters (Controllers, Repos)
    ↑
Frameworks & Drivers (Web, DB)
```

### Dependency Rule

**Dependencies point inward only!**

- Outer layers depend on inner layers
- Inner layers know nothing about outer layers
- Business logic independent of frameworks

### Implementation Guidelines

**1. Domain Layer (Entities):**
```go
// internal/flight/domain/entity/flight.go
package entity

type Flight struct {
    id            string
    flightNumber  FlightNumber  // Value object
    status        FlightStatus
}

// Business logic in domain
func (f *Flight) Delay(duration time.Duration) error {
    if !f.CanBeDelayed() {
        return ErrCannotDelay
    }
    f.departureTime = f.departureTime.Add(duration)
    f.status = FlightStatusDelayed
    return nil
}
```

**2. Use Cases:**
```go
// internal/flight/usecase/delay_flight.go
package usecase

type DelayFlightUseCase struct {
    flightRepo repository.FlightRepository // Interface
}

func (uc *DelayFlightUseCase) Execute(ctx context.Context, input Input) error {
    flight, _ := uc.flightRepo.FindByID(ctx, input.FlightID)
    flight.Delay(input.Duration)
    return uc.flightRepo.Save(ctx, flight)
}
```

**3. Repositories (Interfaces in Domain, Implementations in Adapters):**
```go
// internal/flight/domain/repository/flight_repository.go
type FlightRepository interface {
    Save(ctx context.Context, flight *entity.Flight) error
    FindByID(ctx context.Context, id string) (*entity.Flight, error)
}

// internal/flight/adapter/repository/postgres_flight_repo.go
type PostgresFlightRepository struct {
    db *sql.DB
}

func (r *PostgresFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    // Implementation
}
```

### Benefits

- ✅ Framework independence
- ✅ Testability (mock everything)
- ✅ Database independence
- ✅ Easy to swap implementations

**Source References:** Three Dots Labs, bxcodec/go-clean-arch (GitHub), Medium (Clean Architecture in Go)

---

## 4. Domain-Driven Design Patterns

### Ubiquitous Language

Use domain terms consistently everywhere:
- Code: `flight.Delay()` not `flight.Postpone()`
- API: `/flights/{id}/delay` not `/flights/{id}/late`
- Documentation: Same terminology

### Building Blocks

**1. Entities (Identity Matters):**
```go
type Flight struct {
    id FlightID  // Unique identity
    // ...
}

func (f *Flight) Equals(other *Flight) bool {
    return f.id == other.id  // Compare by identity
}
```

**2. Value Objects (Immutable, No Identity):**
```go
type Money struct {
    amount   decimal.Decimal
    currency Currency
}

// Immutable - returns new instance
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{m.amount.Add(other.amount), m.currency}, nil
}
```

**3. Aggregates (Transactional Boundaries):**
```go
type Booking struct {  // Aggregate Root
    id         BookingID
    passengers []Passenger  // Internal entities
    status     BookingStatus
}

// Only root methods modify aggregate
func (b *Booking) AddPassenger(info PassengerInfo) error {
    // Enforce invariants
    if len(b.passengers) >= MaxPassengers {
        return ErrTooManyPassengers
    }
    b.passengers = append(b.passengers, NewPassenger(info))
    return nil
}
```

**4. Domain Events:**
```go
type FlightDelayed struct {
    eventID      string
    occurredAt   time.Time
    flightID     string
    delayDuration time.Duration
}
```

**5. Repositories (One Per Aggregate Root):**
```go
type BookingRepository interface {
    Save(ctx context.Context, booking *Booking) error
    FindByID(ctx context.Context, id BookingID) (*Booking, error)
}
```

### Strategic Patterns

- **Bounded Contexts**: Flight Context, Passenger Context, Payment Context
- **Context Mapping**: Define relationships between contexts
- **Anti-Corruption Layer**: Translate between contexts

**Source References:** Three Dots Labs (DDD Lite in Go), Programming Percy, Mario Carrion, SayOne Tech

---

## 5. SOLID Principles in Go

### Single Responsibility (SRP)

✅ **Good:**
```go
type FlightRepository struct { db *sql.DB }  // Data access only
type NotificationService struct { emailClient EmailClient }  // Notifications only
type PricingService struct { rules PricingRules }  // Pricing only
```

❌ **Bad:**
```go
type FlightService struct {  // Too many responsibilities!
    // Database, email, pricing, reports all in one
}
```

### Open/Closed (OCP)

✅ **Good** (use interfaces):
```go
type PaymentMethod interface {
    ProcessPayment(ctx context.Context, amount Money) error
}

type CreditCardPayment struct {}
type PayPalPayment struct {}
type CryptoPayment struct {}  // Added without modifying existing code

type PaymentProcessor struct {
    methods map[string]PaymentMethod
}

func (p *PaymentProcessor) RegisterMethod(method PaymentMethod) {
    p.methods[method.GetMethodName()] = method
}
```

### Liskov Substitution (LSP)

Subtypes must be substitutable for base types.

✅ **Good:**
```go
type Flight interface {
    GetDuration() time.Duration
    GetDistance() float64
}

// Both implementations honor the contract
type RegularFlight struct {}
type CharterFlight struct {}
```

### Interface Segregation (ISP)

Many small interfaces > one large interface.

✅ **Good:**
```go
type FlightReader interface {
    GetFlight(ctx context.Context, id string) (*Flight, error)
}

type FlightStatusUpdater interface {
    UpdateStatus(ctx context.Context, id string, status FlightStatus) error
}

type FlightScheduler interface {
    ScheduleFlight(ctx context.Context, flight *Flight) error
}
```

### Dependency Inversion (DIP)

Depend on abstractions, not concretions.

✅ **Good:**
```go
type CreateFlightUseCase struct {
    flightRepo repository.FlightRepository  // Interface!
    eventBus   event.EventBus                // Interface!
}

func NewCreateFlightUseCase(
    repo repository.FlightRepository,
    bus event.EventBus,
) *CreateFlightUseCase {
    return &CreateFlightUseCase{flightRepo: repo, eventBus: bus}
}
```

**Source References:** Dave Cheney (SOLID Go Design), Medium (Understanding SOLID in Golang), PackageMain.tech

---

## 6. Test-Driven Development

### TDD Cycle

```
🔴 RED    → Write failing test
🟢 GREEN  → Make it pass (minimal code)
🔵 REFACTOR → Improve quality
🔁 REPEAT
```

### Testing Pyramid

```
     E2E (10%)        ← Few, slow, expensive
 Integration (20%)    ← Some, moderate speed
   Unit (70%)         ← Many, fast, cheap
```

### Test Patterns

**Table-Driven Tests:**
```go
func TestFlightNumber_Validation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "AA123", false},
        {"too short", "A1", true},
        {"no airline code", "123", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewFlightNumber(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Mocking Frameworks

**testify/mock vs GoMock:**
- **Testify**: User-friendlier, more actively maintained, better error messages
- **GoMock**: More type-safe, official Google library

**Recommended:** testify/mock with mockery CLI

### Testing Tools

- **Unit Tests:** `testing` package, `testify`
- **Integration Tests:** `testcontainers-go`
- **Mocking:** `testify/mock`, `mockery`
- **API Tests:** `httptest`
- **Load Tests:** `k6`, `vegeta`
- **Race Detector:** `go test -race`

**Source References:** Packt (Test-Driven Development in Go), JetBrains (Go Testing Best Practices), testify GitHub

---

## 7. API Design & Versioning

### REST API Best Practices

**Versioning Strategies:**

1. **URL Path (Recommended):**
   - `/api/v1/flights`
   - `/api/v2/flights`
   - ✅ Simple, explicit, widely used

2. **Header-Based:**
   - `Accept-Version: v1`
   - ✅ Cleaner URLs, more flexible

3. **Content Negotiation:**
   - `Accept: application/vnd.company.v1+json`
   - ✅ Granular control

### When to Version

- ✅ Breaking changes → New major version
- ❌ New endpoints/fields → No version change

### OpenAPI/Swagger

**Tools:**
- **swag** (Code-first): Annotations → Swagger
- **go-swagger** (OpenAPI 2.0): Spec → Code
- **oapi-codegen** (OpenAPI 3.0): Spec → Code

**Recommended:** swag for code-first approach

```go
// @Summary Delay a flight
// @Description Delays a flight by specified duration
// @Tags flights
// @Accept json
// @Produce json
// @Param id path string true "Flight ID"
// @Param request body DelayRequest true "Delay details"
// @Success 200 {object} FlightResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/flights/{id}/delay [put]
func (h *FlightHandler) DelayFlight(c *gin.Context) {
    // Implementation
}
```

### Pagination

**Cursor-Based (Recommended for Real-time Data):**
```go
GET /api/v1/flights?cursor=abc123&limit=20
Response:
{
    "data": [...],
    "next_cursor": "xyz789"
}
```

**Benefits:**
- 10x faster than offset
- Handles data mutations gracefully
- No duplicate/missed records

**Offset-Based (For Static Data):**
```go
GET /api/v1/flights?offset=20&limit=20
```

### Idempotency

For POST requests, use idempotency keys:
```go
POST /api/v1/bookings
Headers:
Idempotency-Key: uuid-here
```

**Implementation:**
- Store key + response in cache/database
- Return cached response for duplicate requests
- Expire keys after 24 hours

**Source References:** Mario Carrion, StackOverflow, RestfulAPI.net, Bun (Cursor Pagination), Zuplo (Idempotency Keys)

---

## 8. Communication Patterns

### Synchronous: gRPC

**When to Use:**
- Internal service-to-service communication
- Need low latency
- Binary data transfer
- Streaming support

**Advantages:**
- 5-8x faster than REST+JSON
- HTTP/2 multiplexing
- Built-in load balancing
- Streaming (unary, server, client, bidirectional)

**Implementation:**
```protobuf
syntax = "proto3";

service FlightService {
    rpc GetFlight(FlightRequest) returns (FlightResponse);
    rpc ListFlights(ListRequest) returns (stream FlightResponse);
}
```

### Asynchronous: Event-Driven (RabbitMQ)

**Exchange Types:**
- **Fanout**: Broadcast to all queues
- **Direct**: Route by routing key
- **Topic**: Route by pattern

**Patterns:**
1. **Event Notification**: "FlightDelayed" event
2. **Event-Carried State Transfer**: Event contains full state
3. **Event Sourcing**: Store events, rebuild state
4. **CQRS**: Separate read/write models

**Best Practices:**
- Use prefixes for keys: `flight:delayed:AA123`
- Implement dead letter queues
- Set message TTL
- Use idempotent consumers

**Source References:** Medium (Event-Driven with RabbitMQ), DEV Community, Three Dots Labs

---

## 9. Data Management

### Database Connection Pooling

**PostgreSQL with pgx:**

```go
config, _ := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))

// Configuration
config.MaxConns = 25  // Max open connections
config.MinConns = 5   // Min idle connections
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute

pool, _ := pgxpool.NewWith Pool(context.Background(), config)
```

**Best Practices:**
- Initialize pool once at startup
- `MaxConns` ≈ number of CPU cores × 2
- Set `MaxConnLifetime` for connection recycling
- Use `MinConns` for ready connections

**Prefer pgx over database/sql** for PostgreSQL.

### Redis Caching

**Cache Patterns:**

1. **Cache-Aside:**
   ```go
   val, err := cache.Get(key)
   if err != nil {
       val = db.Query()
       cache.Set(key, val, ttl)
   }
   ```

2. **Write-Through:**
   ```go
   db.Save(data)
   cache.Set(key, data, ttl)
   ```

**Key Naming:**
- Use colons for hierarchy: `flight:AA123:status`
- Prefix by service: `booking:PNR123:passenger:1`

**TTL Strategy:**
- Static data: 24h
- Dynamic data: 30s-5m
- User sessions: 30m (sliding)

### Database Migrations

**golang-migrate:**

```bash
# Create migration
migrate create -ext sql -dir migrations -seq add_flights_table

# Apply
migrate -path migrations -database "postgres://..." up

# Rollback
migrate -path migrations -database "postgres://..." down 1
```

**Best Practices:**
- Use timestamps for versioning
- Keep migrations small
- Always test before production
- Use transactions
- Write reversible migrations (up + down)

**Source References:** Better Stack (golang-migrate), Medium (Connection Pooling), Redis.io, Atlas (Database Migrations)

---

## 10. Error Handling

### Core Principles

1. **Always check errors explicitly**
2. **Wrap errors with context**
3. **Use custom error types**
4. **Never shadow errors**

### Implementation

**Custom Errors:**
```go
var (
    ErrNotFound       = errors.New("resource not found")
    ErrInvalidInput   = errors.New("invalid input")
    ErrUnauthorized   = errors.New("unauthorized")
)

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

**Error Wrapping (Go 1.13+):**
```go
func (s *Service) GetFlight(id string) (*Flight, error) {
    flight, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get flight %s: %w", id, err)
    }
    return flight, nil
}

// Check wrapped error
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

### Error Types for HTTP

```go
func MapErrorToHTTPStatus(err error) int {
    switch {
    case errors.Is(err, ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, ErrInvalidInput):
        return http.StatusBadRequest
    case errors.Is(err, ErrUnauthorized):
        return http.StatusUnauthorized
    default:
        return http.StatusInternalServerError
    }
}
```

### Security Considerations

- ❌ Don't expose internal errors to clients
- ✅ Log detailed errors server-side
- ✅ Return generic errors to clients

**Source References:** Mario Carrion (Error Handling), Earthly Blog, Medium (Error Handling in Go), JetBrains Guide

---

## 11. Concurrency Patterns

### Core Principle

**"Don't communicate by sharing memory; share memory by communicating."**

### Goroutines & Channels

**Worker Pool:**
```go
func WorkerPool(tasks <-chan Task, results chan<- Result, workers int) {
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range tasks {
                result := process(task)
                results <- result
            }
        }()
    }
    wg.Wait()
    close(results)
}
```

**Fan-Out/Fan-In:**
```go
func FanOut(in <-chan int, workers int) []<-chan int {
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        channels[i] = worker(in)
    }
    return channels
}

func FanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for n := range c {
                out <- n
            }
        }(ch)
    }
    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}
```

### Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

select {
case result := <-processChan:
    return result
case <-ctx.Done():
    return ctx.Err()
}
```

**Best Practices:**
- Always call `cancel()` (use defer)
- Pass context as first parameter
- Check `ctx.Done()` in loops
- Use short timeouts near I/O boundaries

### Race Detection

```bash
go test -race ./...
go build -race
```

**Source References:** O'Reilly (Concurrency in Go), GetStream, FreeCodeCamp, Go.dev (Pipelines)

---

## 12. Observability & Monitoring

### Logging (Structured)

**Recommended: zap or zerolog**

**Performance:**
- zerolog: 380 ns/op, 1 allocation
- zap: 656 ns/op, 5 allocations
- zap sugared: 935 ns/op, 10 allocations

**Implementation:**
```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("booking created",
    zap.String("booking_id", booking.ID),
    zap.String("pnr", booking.PNR),
    zap.Int("passengers", len(booking.Passengers)),
)
```

**Best Practices:**
- Use structured logging (JSON)
- Log actionable information
- Consistent field names
- Appropriate log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- Retention: 30d hot, 90d warm, 1y cold

### Metrics (Prometheus)

**RED Metrics (Request-oriented):**
- **R**ate: Requests per second
- **E**rrors: Error rate %
- **D**uration: Response time (p50, p95, p99)

**USE Metrics (Resource-oriented):**
- **U**tilization: CPU, memory, disk %
- **S**aturation: Queue depth
- **E**rrors: Failed operations

**Implementation:**
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"method", "endpoint", "status"},
    )
)

func init() {
    prometheus.MustRegister(httpDuration)
}
```

**Dashboard Strategy:**
- Service health overview
- Database performance
- Business KPIs
- SLA compliance

### Distributed Tracing (Jaeger + OpenTelemetry)

**Span Structure:**
```
Flight Booking Trace
├─ API Gateway (2ms)
├─ Booking Service (150ms)
│  ├─ Flight Service (20ms)
│  ├─ Passenger Service (30ms)
│  ├─ Payment Service (80ms)
│  └─ Notification Service (async)
└─ Total (152ms)
```

**Implementation:**
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

tracer := otel.Tracer("booking-service")
ctx, span := tracer.Start(ctx, "create-booking")
defer span.End()

span.SetAttributes(
    attribute.String("booking.id", bookingID),
    attribute.Int("passenger.count", passengerCount),
)
```

**Best Practices:**
- 100% sampling for errors
- 10% sampling for normal requests
- Meaningful span names
- Propagate context across services

**Source References:** Better Stack (Logging in Go), SigNoz (Zap Logger), Medium (OpenTelemetry), Komodor (Health Checks)

---

## 13. Security Best Practices

### Authentication & Authorization

**JWT Best Practices:**
- Use asymmetric algorithms (ES256, EdDSA)
- Always validate: algorithm, signature, claims (iat, exp, iss, aud)
- Transmit over HTTPS only
- Short expiration times (15min access, 7d refresh)

**Implementation:**
```go
import "github.com/golang-jwt/jwt/v5"

func ValidateJWT(tokenString string, publicKey *rsa.PublicKey) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate algorithm
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return publicKey, nil
    })

    if err != nil || !token.Valid {
        return nil, err
    }

    return token.Claims.(*Claims), nil
}
```

### Input Validation & Sanitization

**Always validate:**
- Type checking
- Length validation
- Format validation (regex)
- Whitelist > Blacklist

**Example:**
```go
import "github.com/go-playground/validator/v10"

type CreateFlightRequest struct {
    FlightNumber string `validate:"required,len=5"`
    Departure    string `validate:"required,len=3,alpha"`
    Arrival      string `validate:"required,len=3,alpha"`
}

validate := validator.New()
err := validate.Struct(request)
```

**XSS Prevention:**
- Use `html/template` (auto-escaping)
- Sanitize user input before rendering

**SQL Injection Prevention:**
- Use prepared statements
- Use ORM (GORM, sqlc)
- Never concatenate SQL strings

### CORS Configuration

```go
import "github.com/rs/cors"

c := cors.New(cors.Options{
    AllowedOrigins: []string{"https://example.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge: 300,
})

handler := c.Handler(router)
```

**Security Headers:**
```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000")
        next.ServeHTTP(w, r)
    })
}
```

### Secrets Management

- ❌ Never store secrets in code/config
- ✅ Use environment variables
- ✅ Use HashiCorp Vault or AWS Secrets Manager
- ✅ Rotate credentials regularly

### Rate Limiting

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(10, 20)  // 10 req/sec, burst of 20

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Source References:** OWASP (JWT Security), StackHawk (CORS, Input Validation), Medium (12 Security Tips), golang.org/x/time/rate

---

## 14. Performance Optimization

### Profiling with pprof

**CPU Profiling:**
```bash
go test -cpuprofile cpu.prof -bench .
go tool pprof cpu.prof
```

**Memory Profiling:**
```bash
go test -memprofile mem.prof -bench .
go tool pprof mem.prof
```

**Web Interface:**
```bash
go tool pprof -http=":8080" cpu.prof
```

**Types of Profiling:**
- CPU profiling (execution time)
- Heap profiling (memory allocation)
- Goroutine profiling (concurrency)
- Block profiling (synchronization delays)
- Mutex profiling (lock contention)

### Memory Optimization

**Key Techniques:**
1. **Escape Analysis**: Keep allocations on stack
2. **sync.Pool**: Reuse allocations
3. **Minimize Allocations**: Reuse buffers

**Example:**
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) []byte {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()

    // Use buffer
    buf.Write(data)
    return buf.Bytes()
}
```

### Garbage Collection Tuning

**Go 1.19+: GOMEMLIMIT**
```bash
# Set memory limit (leave 5-10% headroom)
GOMEMLIMIT=450MiB ./app  # For 512MB container
```

**GOGC (GC frequency):**
```bash
GOGC=100  # Default: GC when heap grows 100%
GOGC=200  # Less frequent GC, more memory
```

### Benchmarking

```go
func BenchmarkFlightCreation(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = NewFlight("AA123", departure, arrival)
    }
}
```

**Run:**
```bash
go test -bench=. -benchmem
```

**Source References:** Go.dev (pprof), Better Programming (Pprof Examples), Go.dev (GC Guide), VictoriaMetrics

---

## 15. Deployment & Infrastructure

### Docker Multi-Stage Builds

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /app/flight-service ./cmd/flight-service

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/flight-service .

USER appuser

EXPOSE 8081

CMD ["./flight-service"]
```

**Benefits:**
- 95% size reduction (350MB → 13MB)
- Enhanced security (minimal attack surface)
- Faster deployments

### Kubernetes Health Checks

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: flight-service
    livenessProbe:
      httpGet:
        path: /health
        port: 8081
      initialDelaySeconds: 10
      periodSeconds: 10

    readinessProbe:
      httpGet:
        path: /ready
        port: 8081
      initialDelaySeconds: 5
      periodSeconds: 5

    startupProbe:
      httpGet:
        path: /startup
        port: 8081
      failureThreshold: 30
      periodSeconds: 10
```

**Health Check Endpoints:**
```go
// GET /health - Liveness (is app alive?)
func Health(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}

// GET /ready - Readiness (can accept traffic?)
func Ready(w http.ResponseWriter, r *http.Request) {
    if !checkDatabaseConnection() {
        w.WriteHeader(http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
}
```

### Graceful Shutdown

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: router}

    // Listen for signals
    ctx, stop := signal.NotifyContext(context.Background(),
        os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // Wait for signal
    <-ctx.Done()
    log.Println("Shutting down gracefully...")

    // Shutdown with timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Fatal(err)
    }
}
```

### Service Discovery

**Consul (Recommended for VMs):**
- Service registration & discovery
- Health checking
- Key-value store
- Distributed configuration

**etcd (Recommended for Kubernetes):**
- Distributed key-value store
- Built-in with Kubernetes
- Watch mechanism for changes

### Circuit Breaker

```go
import "github.com/afex/hystrix-go/hystrix"

hystrix.ConfigureCommand("get_flight", hystrix.CommandConfig{
    Timeout:                1000,  // 1 second
    MaxConcurrentRequests:  100,
    ErrorPercentThreshold:  50,
    RequestVolumeThreshold: 10,
    SleepWindow:            5000,  // 5 seconds
})

err := hystrix.Do("get_flight", func() error {
    return fetchFlightFromService()
}, func(err error) error {
    // Fallback
    return getCachedFlight()
})
```

### Retry with Exponential Backoff

```go
import "github.com/sethvargo/go-retry"

err := retry.Do(ctx, retry.WithMaxRetries(3, retry.NewExponential(1*time.Second)), func(ctx context.Context) error {
    return callExternalAPI()
})
```

**Source References:** Medium (Multi-Stage Builds), Komodor (K8s Health Checks), Mario Carrion (Graceful Shutdown), Leapcell (Circuit Breakers), GitHub (go-retry)

---

## Summary: Implementation Checklist

### Must-Have Practices

- [ ] **Project Structure**: Follow golang-standards/project-layout
- [ ] **Clean Architecture**: 4-layer structure with dependency inversion
- [ ] **DDD**: Rich domain models, aggregates, value objects
- [ ] **SOLID**: Apply all 5 principles
- [ ] **TDD**: Red-Green-Refactor, 70% unit tests
- [ ] **Error Handling**: Wrap errors, custom types, structured logging
- [ ] **Context**: Use for cancellation, timeouts
- [ ] **Structured Logging**: zap or zerolog
- [ ] **Metrics**: Prometheus with RED & USE metrics
- [ ] **Tracing**: Jaeger with OpenTelemetry
- [ ] **Security**: JWT validation, input sanitization, CORS, rate limiting
- [ ] **Database**: Connection pooling, migrations (golang-migrate)
- [ ] **Caching**: Redis with proper key naming, TTLs
- [ ] **Health Checks**: Liveness, readiness, startup probes
- [ ] **Graceful Shutdown**: Handle SIGTERM properly
- [ ] **Docker**: Multi-stage builds, non-root user
- [ ] **Testing**: Unit (testify), integration (testcontainers), E2E
- [ ] **Linting**: golangci-lint with strict configuration
- [ ] **Documentation**: OpenAPI/Swagger
- [ ] **Resilience**: Circuit breakers, retries, timeouts
- [ ] **Observability**: Full stack (logs, metrics, traces)

---

## Research Sources Summary

This document synthesizes research from **40+ authoritative sources** including:

- Official Go documentation (go.dev, pkg.go.dev)
- Industry leaders (JetBrains, Google, Uber, Netflix, HashiCorp)
- Community resources (Medium, DEV.to, GitHub)
- Academic publications (O'Reilly, Packt)
- Technology blogs (Three Dots Labs, Better Stack, LogRocket)
- OWASP security guidelines
- Kubernetes & CNCF documentation

**Last Updated**: 2025-11-19
**Research Period**: 2024-2025
**Applicable to**: Go 1.21+

---

**All practices in this document are mandatory for Airport Services platform development.**

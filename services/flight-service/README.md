# Flight Management Service

A production-grade microservice for managing flight operations built with Clean Architecture, Domain-Driven Design (DDD), SOLID principles, and Test-Driven Development (TDD).

## Features

- ✅ **Create Flight** - Schedule new flights
- ✅ **Get Flight** - Retrieve flight details by ID
- ✅ **Update Status** - Change flight status (scheduled → boarding → departed → in_flight → landed → arrived)
- ✅ **Delay Flight** - Delay a flight with reason
- ✅ **Cancel Flight** - Cancel a flight with reason
- ✅ **Status Validation** - Enforces valid state transitions
- ✅ **Business Rules** - Cannot delay/cancel after departure

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│  Interfaces (HTTP/gRPC)                                 │
│  - Handlers, DTOs, Routes                               │
├─────────────────────────────────────────────────────────┤
│  Use Cases (Application Business Rules)                 │
│  - CreateFlight, UpdateStatus, Delay, Cancel            │
├─────────────────────────────────────────────────────────┤
│  Domain (Enterprise Business Rules)                     │
│  - Entities, Value Objects, Repository Interfaces       │
├─────────────────────────────────────────────────────────┤
│  Infrastructure (Frameworks & Drivers)                  │
│  - PostgreSQL, Redis, RabbitMQ                          │
└─────────────────────────────────────────────────────────┘
```

### Project Structure

```
services/flight-service/
├── cmd/
│   └── main.go                  # Application entry point
├── internal/
│   ├── domain/                  # Domain layer
│   │   ├── entity/             # Flight aggregate root
│   │   ├── valueobject/        # FlightNumber, FlightStatus
│   │   └── repository/         # Repository interfaces
│   ├── usecase/                # Use case layer
│   │   ├── create_flight.go
│   │   ├── get_flight.go
│   │   ├── update_flight_status.go
│   │   ├── delay_flight.go
│   │   └── cancel_flight.go
│   ├── infrastructure/          # Infrastructure layer
│   │   └── persistence/
│   │       └── postgres/       # PostgreSQL repository implementation
│   └── interfaces/             # Interface adapters
│       └── http/
│           ├── handler/        # HTTP handlers
│           ├── dto/            # Request/Response DTOs
│           └── router/         # Route configuration
├── migrations/                  # Database migrations
├── deployments/
│   └── kubernetes/             # K8s manifests
├── Dockerfile                   # Multi-stage Docker build
└── .env.example                # Environment variables template
```

## Tech Stack

- **Go 1.24+** - Programming language
- **Gin** - HTTP framework
- **PostgreSQL** - Primary database (pgx driver)
- **Redis** - Caching layer
- **RabbitMQ** - Event-driven messaging
- **Zap** - Structured logging
- **Docker** - Containerization
- **Kubernetes** - Orchestration

## Prerequisites

- Go 1.24+
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+ (optional for events)
- Docker & Docker Compose (for local development)

## Quick Start

### 1. Clone and Setup

```bash
cd services/flight-service
cp .env.example .env
# Edit .env with your configuration
```

### 2. Start Dependencies (Docker Compose)

```bash
# From project root
docker-compose up -d postgres redis rabbitmq
```

### 3. Run Database Migrations

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/flight_db?sslmode=disable" up
```

### 4. Run the Service

```bash
# Development mode
go run cmd/main.go

# Or build and run
go build -o flight-service cmd/main.go
./flight-service
```

The service will start on `http://localhost:8080`

## API Endpoints

### Health Checks

```http
GET /health       # Liveness probe
GET /ready        # Readiness probe
```

### Flight Operations

```http
POST   /api/v1/flights              # Create a flight
GET    /api/v1/flights/:id          # Get flight by ID
PATCH  /api/v1/flights/:id/status   # Update flight status
POST   /api/v1/flights/:id/delay    # Delay a flight
POST   /api/v1/flights/:id/cancel   # Cancel a flight
```

### Example: Create Flight

```bash
curl -X POST http://localhost:8080/api/v1/flights \
  -H "Content-Type: application/json" \
  -d '{
    "flight_number": "AA123",
    "origin": "JFK",
    "destination": "LAX",
    "departure_time": "2025-12-01T10:00:00Z",
    "arrival_time": "2025-12-01T13:30:00Z",
    "aircraft_type": "Boeing 737",
    "gate": "A12"
  }'
```

### Example: Update Status

```bash
curl -X PATCH http://localhost:8080/api/v1/flights/{id}/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "boarding"
  }'
```

### Example: Delay Flight

```bash
curl -X POST http://localhost:8080/api/v1/flights/{id}/delay \
  -H "Content-Type: application/json" \
  -d '{
    "delay_minutes": 30,
    "reason": "Weather conditions"
  }'
```

## Testing

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/domain/entity/... -v
go test ./internal/usecase/... -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Current Test Coverage**: 40+ tests, 0 failures

## Database Schema

The service uses PostgreSQL with the following schema:

- **flights** table with strategic indexes for performance
- Constraints for data integrity (valid statuses, IATA codes, arrival after departure)
- Auto-updating timestamps via triggers

See `migrations/000001_create_flights_table.up.sql` for full schema.

## Docker

### Build Image

```bash
docker build -t flight-service:latest -f services/flight-service/Dockerfile .
```

### Run Container

```bash
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PASSWORD=postgres \
  flight-service:latest
```

## Kubernetes Deployment

```bash
# Create namespace
kubectl create namespace airport-services

# Apply manifests
kubectl apply -f deployments/kubernetes/

# Check status
kubectl get pods -n airport-services
kubectl logs -f deployment/flight-service -n airport-services
```

## Environment Variables

See `.env.example` for all available configuration options.

### Required Variables

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database connection
- `JWT_SECRET` - JWT signing key (production)

### Optional Variables

- `REDIS_HOST`, `REDIS_PORT` - Redis caching
- `RABBITMQ_URL` - Message queue
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `TRACING_ENABLED` - Enable distributed tracing

## Architecture Principles

### Clean Architecture ✅
- Clear separation of concerns
- Domain layer independent of frameworks
- Dependency Inversion Principle

### Domain-Driven Design ✅
- Rich domain models (Flight entity)
- Value objects (FlightNumber, FlightStatus)
- Aggregates with invariants
- Repository pattern

### SOLID Principles ✅
- **S**ingle Responsibility - Each layer has one reason to change
- **O**pen/Closed - Status transitions extensible without modification
- **L**iskov Substitution - Proper use of interfaces
- **I**nterface Segregation - Small, focused interfaces
- **D**ependency Inversion - Use cases depend on abstractions

### Test-Driven Development ✅
- All code written test-first (Red-Green-Refactor)
- Unit tests for domain logic
- Integration tests for repositories
- HTTP handler tests

## Performance

- **Connection Pooling**: pgx with configurable pool size
- **Indexed Queries**: 8 strategic indexes on flights table
- **Pagination**: All list endpoints support pagination
- **Rate Limiting**: 100 req/s with burst of 200
- **Caching**: Redis integration for hot data

## Security

- ✅ Non-root container user
- ✅ Security headers (X-Frame-Options, CSP, etc.)
- ✅ CORS configuration
- ✅ Rate limiting per IP
- ✅ Input validation
- ✅ SQL injection prevention (parameterized queries)
- ✅ JWT authentication ready

## Monitoring

- **Logs**: Structured JSON logs with Zap
- **Metrics**: Prometheus metrics on `:9090/metrics`
- **Tracing**: Jaeger/OpenTelemetry integration
- **Health**: `/health` and `/ready` endpoints

## Contributing

This service follows strict architectural guidelines:

1. Write tests first (TDD)
2. Keep domain layer pure (no frameworks)
3. Use cases orchestrate, don't contain business logic
4. All business rules in domain entities
5. Follow existing patterns and structure

## License

Part of the Airport Services ecosystem.

# Flight Management Service

A production-grade microservice for managing flight operations built with Clean Architecture, Domain-Driven Design (DDD), SOLID principles, and Test-Driven Development (TDD).

## Features

### Flight Operations
- ✅ **Create Flight** - Schedule new flights
- ✅ **Get Flight** - Retrieve flight details by ID
- ✅ **List Flights** - Query flights with 7+ filter types (status, origin, destination, date range, etc.)
- ✅ **Search by Number** - Find flights by flight number with optional date filtering
- ✅ **Update Status** - Change flight status (scheduled → boarding → departed → in_flight → landed → arrived)
- ✅ **Delay Flight** - Delay a flight with reason
- ✅ **Cancel Flight** - Cancel a flight with reason

### Architecture Features
- ✅ **Status Validation** - Enforces valid state transitions
- ✅ **Business Rules** - Cannot delay/cancel after departure
- ✅ **Event-Driven** - Publishes domain events to RabbitMQ for inter-service communication
- ✅ **Async Communication** - 7 event types (FlightCreated, StatusChanged, Delayed, Cancelled, BoardingStarted, Departed, Arrived)

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
POST   /api/v1/flights                        # Create a flight
GET    /api/v1/flights                        # List flights with filters
GET    /api/v1/flights/search?number=AA123    # Search by flight number
GET    /api/v1/flights/:id                    # Get flight by ID
PATCH  /api/v1/flights/:id/status             # Update flight status
POST   /api/v1/flights/:id/delay              # Delay a flight
POST   /api/v1/flights/:id/cancel             # Cancel a flight
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

### Example: List Flights with Filters

```bash
# List all flights (paginated)
curl http://localhost:8080/api/v1/flights?page=1&limit=20

# Filter by status
curl http://localhost:8080/api/v1/flights?status=boarding

# Filter by origin and destination
curl http://localhost:8080/api/v1/flights?origin=JFK&destination=LAX

# Filter by date range
curl "http://localhost:8080/api/v1/flights?start_date=2025-12-01T00:00:00Z&end_date=2025-12-31T23:59:59Z"

# Get only delayed flights
curl http://localhost:8080/api/v1/flights?delayed_only=true
```

### Example: Search by Flight Number

```bash
# Search for flight AA123 (latest)
curl http://localhost:8080/api/v1/flights/search?number=AA123

# Search for specific date
curl http://localhost:8080/api/v1/flights/search?number=AA123&date=2025-12-01
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

## Event-Driven Architecture

The Flight Service publishes domain events to RabbitMQ for asynchronous inter-service communication.

### Published Events

The service publishes 7 types of domain events:

1. **flight.created** - When a new flight is scheduled
2. **flight.status.changed** - When flight status is updated
3. **flight.delayed** - When a flight is delayed
4. **flight.cancelled** - When a flight is cancelled
5. **flight.boarding.started** - When boarding begins
6. **flight.departed** - When flight departs
7. **flight.arrived** - When flight arrives at destination

### Event Routing

- Uses RabbitMQ **topic exchange** for flexible routing
- Events are published with routing key = event type (e.g., `flight.created`)
- Other services can subscribe to specific event types using routing patterns
- Messages are **persistent** for reliability

### Configuration

Events are optional and don't block operations:
```bash
# Enable events (requires RabbitMQ)
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=flight-events

# Service works without RabbitMQ (events disabled)
# Simply omit RABBITMQ_URL environment variable
```

### Example Event Payload

```json
{
  "flight_id": "550e8400-e29b-41d4-a716-446655440000",
  "flight_number": "AA123",
  "origin": "JFK",
  "destination": "LAX",
  "delay_duration": 30,
  "reason": "Weather conditions",
  "new_departure_time": "2025-12-01T10:30:00Z",
  "new_arrival_time": "2025-12-01T14:00:00Z",
  "delayed_at": "2025-12-01T09:45:00Z"
}
```

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

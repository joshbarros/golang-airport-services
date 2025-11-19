# Airport Services - Quick Start Guide

Complete microservices ecosystem for airport operations. This guide will get you up and running in minutes.

## What's Included

### Microservices
- **Flight Service** (Port 8080) - Flight management and operations
- **Notification Service** (Port 8081) - Multi-channel notifications (Email, SMS, Push)

### Infrastructure
- **PostgreSQL** (Port 5432) - Primary database
- **Redis** (Port 6379) - Caching layer
- **RabbitMQ** (Ports 5672, 15672) - Message broker for event-driven architecture
- **MailHog** (Ports 1025, 8025) - Email testing

### Monitoring & Management
- **Prometheus** (Port 9090) - Metrics collection
- **Grafana** (Port 3000) - Metrics visualization
- **Jaeger** (Port 16686) - Distributed tracing
- **Elasticsearch** (Port 9200) - Search and analytics
- **Kibana** (Port 5601) - Elasticsearch UI
- **pgAdmin** (Port 5050) - PostgreSQL admin
- **MinIO** (Ports 9000, 9001) - S3-compatible object storage
- **Consul** (Port 8500) - Service discovery

## Prerequisites

- Docker & Docker Compose installed
- 8GB+ RAM recommended
- Ports 8080, 8081, 5432, 5672, 15672 available

## Quick Start (5 minutes)

### 1. Start Core Infrastructure

```bash
# Start databases and message broker
docker-compose up -d postgres redis rabbitmq mailhog

# Check status
docker-compose ps

# Wait for health checks (30 seconds)
watch docker-compose ps
```

### 2. Build and Start Microservices

```bash
# Build images (first time only, ~2 minutes)
docker-compose build flight-service notification-service

# Start services
docker-compose up -d flight-service notification-service

# Watch logs
docker-compose logs -f flight-service notification-service
```

### 3. Verify Services are Running

```bash
# Check health endpoints
curl http://localhost:8080/health  # Flight Service
curl http://localhost:8081/health  # Notification Service

# Check RabbitMQ Management UI
open http://localhost:15672  # user: airport, password: airport123
```

## Demo: Event-Driven Notifications

### Step 1: Create a Flight

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

Expected response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "flight_number": "AA123",
  "status": "scheduled",
  ...
}
```

### Step 2: Delay the Flight (Triggers Notification)

```bash
# Save the flight ID from previous response
FLIGHT_ID="<your-flight-id>"

curl -X POST http://localhost:8080/api/v1/flights/$FLIGHT_ID/delay \
  -H "Content-Type: application/json" \
  -d '{
    "delay_minutes": 30,
    "reason": "Weather conditions"
  }'
```

### Step 3: Watch the Magic Happen!

```bash
# Terminal 1: Watch Flight Service logs
docker-compose logs -f flight-service

# Terminal 2: Watch Notification Service logs
docker-compose logs -f notification-service

# Terminal 3: Check MailHog UI for email
open http://localhost:8025
```

You should see:
1. **Flight Service** publishes `flight.delayed` event to RabbitMQ
2. **Notification Service** consumes event and sends email
3. **MailHog** receives the email notification

**Email Preview:**
```
Subject: Flight AA123 Delayed
Body: Flight AA123 has been delayed by 30 minutes.
      Reason: Weather conditions.
      New departure time: 10:30
```

## API Examples

### List All Flights

```bash
curl http://localhost:8080/api/v1/flights
```

### Search by Flight Number

```bash
curl "http://localhost:8080/api/v1/flights/search?number=AA123"
```

### Filter by Status

```bash
curl "http://localhost:8080/api/v1/flights?status=delayed"
```

### Cancel a Flight

```bash
curl -X POST http://localhost:8080/api/v1/flights/$FLIGHT_ID/cancel \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Mechanical issue"
  }'
```

## Management UIs

Access these web interfaces:

- **RabbitMQ Management**: http://localhost:15672 (airport/airport123)
- **MailHog**: http://localhost:8025 (see sent emails)
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686 (distributed tracing)
- **pgAdmin**: http://localhost:5050 (admin@airport.com/admin)
- **Kibana**: http://localhost:5601
- **MinIO**: http://localhost:9001 (airport/airport123)
- **Consul**: http://localhost:8500

## Architecture Highlights

### Event-Driven Communication

```
┌──────────────────┐                    ┌──────────────────────┐
│  Flight Service  │                    │ Notification Service │
│                  │                    │                      │
│  1. Delay Flight │                    │                      │
│  2. Publish Event├────RabbitMQ────────>│  4. Consume Event    │
│     "flight.     │   Topic Exchange   │  5. Send Email       │
│      delayed"    │                    │  6. Send SMS         │
└──────────────────┘                    └──────────────────────┘
         │
         ▼
    PostgreSQL
```

### Clean Architecture

Both services follow:
- **Domain Layer**: Business logic (entities, value objects)
- **Use Case Layer**: Application orchestration
- **Infrastructure Layer**: External integrations (DB, RabbitMQ, SMTP)
- **Interface Layer**: HTTP APIs

## Troubleshooting

### Services won't start

```bash
# Check logs
docker-compose logs flight-service
docker-compose logs notification-service

# Restart services
docker-compose restart flight-service notification-service
```

### Can't connect to database

```bash
# Check PostgreSQL is healthy
docker-compose ps postgres

# View database logs
docker-compose logs postgres

# Manually test connection
docker exec -it airport-postgres psql -U airport -d airport
```

### RabbitMQ issues

```bash
# Check RabbitMQ logs
docker-compose logs rabbitmq

# Access management UI
open http://localhost:15672

# Check exchanges and queues
# Should see "flight-events" exchange
# Should see "notification-service-queue" queue
```

### Notifications not sending

```bash
# Check if event consumer started
docker-compose logs notification-service | grep "Started consuming"

# Check RabbitMQ for messages
open http://localhost:15672
# Navigate to Queues → notification-service-queue
# Check message count

# Check MailHog for emails
open http://localhost:8025
```

## Clean Up

### Stop Services (Keep Data)

```bash
docker-compose stop
```

### Stop and Remove Everything

```bash
docker-compose down

# Also remove volumes (deletes all data)
docker-compose down -v
```

### Remove Images

```bash
docker rmi flight-service notification-service
```

## Next Steps

1. **Add More Services**:
   - Passenger Service
   - Booking Service
   - Check-in Service
   - Baggage Service

2. **Implement gRPC**:
   - Add gRPC alongside HTTP
   - Service-to-service communication

3. **Add Authentication**:
   - JWT tokens
   - OAuth2 integration
   - API Gateway

4. **Production Deployment**:
   - Kubernetes manifests
   - Helm charts
   - CI/CD pipelines

## Documentation

- [Flight Service README](services/flight-service/README.md)
- [Notification Service README](services/notification-service/README.md)
- [Architecture Documentation](docs/architecture.md)
- [Best Practices Research](docs/research/)

## Support

- GitHub Issues: https://github.com/joshbarros/golang-airport-services/issues
- Email: support@airport-services.com

---

**Built with Clean Architecture, Domain-Driven Design, SOLID Principles, and Test-Driven Development** ✨

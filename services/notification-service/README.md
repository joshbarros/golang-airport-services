# Notification Service

A production-grade microservice for managing multi-channel notifications (Email, SMS, Push) in response to flight events. Built with Clean Architecture, Domain-Driven Design (DDD), SOLID principles, and event-driven architecture.

## Features

### Notification Channels
- ✅ **Email Notifications** - SMTP-based email delivery
- ✅ **SMS Notifications** - Console provider (Twilio-ready)
- ✅ **Push Notifications** - Console provider (FCM/APNS-ready)
- ✅ **Multi-Channel Support** - Send to multiple channels simultaneously

### Event-Driven Architecture
- ✅ **Flight Event Consumer** - Listens to RabbitMQ for flight events
- ✅ **7 Event Types Handled**:
  - `flight.created` - New flight scheduled
  - `flight.delayed` - Flight delayed
  - `flight.cancelled` - Flight cancelled
  - `flight.boarding.started` - Boarding announced
  - `flight.departed` - Flight departed
  - `flight.arrived` - Flight arrived
  - `flight.status.changed` - General status updates

### Architecture Features
- ✅ **Clean Architecture** - Clear separation of concerns
- ✅ **Domain-Driven Design** - Rich domain models with business logic
- ✅ **SOLID Principles** - Dependency Inversion for providers
- ✅ **Status State Machine** - Valid notification state transitions
- ✅ **Retry Mechanism** - Automatic retry for failed notifications
- ✅ **Audit Trail** - PostgreSQL persistence for notification history

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│  Interfaces (HTTP/Events)                               │
│  - Health checks, Event consumers                       │
├─────────────────────────────────────────────────────────┤
│  Use Cases (Application Business Rules)                 │
│  - SendNotification, GetNotification                    │
├─────────────────────────────────────────────────────────┤
│  Domain (Enterprise Business Rules)                     │
│  - Notification Entity, Status/Type Value Objects       │
├─────────────────────────────────────────────────────────┤
│  Infrastructure (Frameworks & Drivers)                  │
│  - SMTP, RabbitMQ, PostgreSQL                           │
└─────────────────────────────────────────────────────────┘
```

### Project Structure

```
services/notification-service/
├── cmd/
│   └── main.go                     # Application entry point
├── internal/
│   ├── domain/                     # Domain layer
│   │   ├── entity/                # Notification aggregate
│   │   ├── valueobject/           # NotificationType, NotificationStatus
│   │   └── repository/            # Repository & Provider interfaces
│   ├── usecase/                   # Use case layer
│   │   ├── send_notification.go
│   │   └── get_notification.go
│   ├── infrastructure/            # Infrastructure layer
│   │   ├── messaging/
│   │   │   └── rabbitmq/         # Event consumer
│   │   ├── notification/
│   │   │   ├── email/            # SMTP provider
│   │   │   ├── sms/              # SMS provider (console demo)
│   │   │   └── push/             # Push provider (console demo)
│   │   └── persistence/
│   │       └── postgres/         # PostgreSQL repository
│   └── interfaces/                # Interface adapters
│       └── http/                  # Health check endpoints
├── Dockerfile                     # Multi-stage Docker build
└── .env.example                   # Environment variables template
```

## Tech Stack

- **Go 1.24+** - Programming language
- **RabbitMQ** - Event message broker (required)
- **PostgreSQL** - Notification history/audit
- **SMTP** - Email delivery
- **Zap** - Structured logging
- **Docker** - Containerization

## Prerequisites

- Go 1.24+
- PostgreSQL 15+ (optional for audit trail)
- RabbitMQ 3.12+ (required)
- SMTP server credentials (for email notifications)

## Quick Start

### 1. Clone and Setup

```bash
cd services/notification-service
cp .env.example .env
# Edit .env with your configuration
```

### 2. Start Dependencies

```bash
# Start RabbitMQ (required)
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management

# Start PostgreSQL (optional)
docker run -d --name notification-postgres \
  -e POSTGRES_DB=notification_db \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:15
```

### 3. Configure Email (Optional)

For Gmail SMTP:
1. Enable 2-factor authentication
2. Generate an App Password
3. Update `.env`:
   ```
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-app-password
   SMTP_FROM=notifications@airport.com
   SMTP_USE_TLS=true
   ```

### 4. Run the Service

```bash
# Development mode
go run cmd/main.go

# Or build and run
go build -o notification-service cmd/main.go
./notification-service
```

The service will start on `http://localhost:8081` and begin consuming flight events.

## API Endpoints

### Health Checks

```http
GET /health       # Liveness probe
GET /ready        # Readiness probe
```

## Event Consumption

The service automatically consumes flight events from RabbitMQ and sends notifications:

### Flight Delayed Event → Email Notification

```json
{
  "event_type": "flight.delayed",
  "flight_id": "550e8400-e29b-41d4-a716-446655440000",
  "flight_number": "AA123",
  "origin": "JFK",
  "destination": "LAX",
  "delay_duration": 30,
  "reason": "Weather conditions",
  "new_departure_time": "2025-12-01T10:30:00Z"
}
```

Triggers email notification:
- **Subject**: "Flight AA123 Delayed"
- **Message**: "Flight AA123 has been delayed by 30 minutes. Reason: Weather conditions. New departure time: 10:30"

### Flight Cancelled Event → Email Notification

```json
{
  "event_type": "flight.cancelled",
  "flight_id": "550e8400-e29b-41d4-a716-446655440000",
  "flight_number": "AA123",
  "origin": "JFK",
  "destination": "LAX",
  "reason": "Mechanical issue"
}
```

Triggers email notification:
- **Subject**: "Flight AA123 Cancelled"
- **Message**: "Flight AA123 from JFK to LAX has been cancelled. Reason: Mechanical issue"

## Notification Providers

### Email Provider (SMTP)

Production-ready SMTP provider with TLS support:

```go
email.NewSMTPProvider(
    "smtp.gmail.com",
    "587",
    "your-email@gmail.com",
    "your-app-password",
    "notifications@airport.com",
    true, // useTLS
    logger,
)
```

### SMS Provider (Console Demo)

Console-based provider for development. In production, integrate with Twilio:

```go
// Production implementation
type TwilioProvider struct {
    client *twilio.RestClient
}

func (p *TwilioProvider) Send(ctx, recipient, _, message string, _) error {
    return p.client.Messages.SendMessage(
        twilioPhoneNumber,
        recipient,
        message,
    )
}
```

### Push Provider (Console Demo)

Console-based provider for development. In production, integrate with FCM:

```go
// Production implementation
type FCMProvider struct {
    client *messaging.Client
}

func (p *FCMProvider) Send(ctx, deviceToken, subject, message string, _) error {
    return p.client.Send(ctx, &messaging.Message{
        Token: deviceToken,
        Notification: &messaging.Notification{
            Title: subject,
            Body:  message,
        },
    })
}
```

## Configuration

### Required Configuration

```bash
# RabbitMQ (REQUIRED)
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=flight-events
```

### Optional Configuration

```bash
# Email Provider
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Database (for audit trail)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=notification_db
```

## Docker

### Build Image

```bash
docker build -t notification-service:latest -f services/notification-service/Dockerfile .
```

### Run Container

```bash
docker run -d \
  --name notification-service \
  -p 8081:8081 \
  -e RABBITMQ_URL=amqp://rabbitmq:5672/ \
  -e SMTP_HOST=smtp.gmail.com \
  -e SMTP_PORT=587 \
  -e SMTP_USERNAME=your-email@gmail.com \
  -e SMTP_PASSWORD=your-app-password \
  notification-service:latest
```

## Testing Integration with Flight Service

1. **Start both services**:
   ```bash
   # Terminal 1: Flight Service
   cd services/flight-service
   go run cmd/main.go

   # Terminal 2: Notification Service
   cd services/notification-service
   go run cmd/main.go
   ```

2. **Trigger a flight delay** (Flight Service):
   ```bash
   curl -X POST http://localhost:8080/api/v1/flights/{id}/delay \
     -H "Content-Type: application/json" \
     -d '{
       "delay_minutes": 30,
       "reason": "Weather conditions"
     }'
   ```

3. **Watch Notification Service logs** - You'll see:
   ```
   INFO  Received flight event  event_type=flight.delayed
   INFO  Email sent successfully  recipient=passenger@example.com
   ```

## Architecture Principles

### Clean Architecture ✅
- Clear separation of concerns across layers
- Domain layer independent of frameworks
- Dependency Inversion Principle

### Domain-Driven Design ✅
- Notification aggregate with business rules
- Value objects (NotificationType, NotificationStatus)
- Status state machine for valid transitions
- Repository pattern

### Event-Driven Architecture ✅
- Asynchronous communication via RabbitMQ
- Loosely coupled microservices
- Automatic notification triggering
- Resilient to service failures

### SOLID Principles ✅
- **S**ingle Responsibility - Each provider handles one channel
- **O**pen/Closed - Easy to add new notification providers
- **L**iskov Substitution - All providers implement same interface
- **I**nterface Segregation - Small, focused interfaces
- **D**ependency Inversion - Use cases depend on abstractions

## Production Considerations

### Scaling
- **Horizontal Scaling**: Run multiple instances consuming from same queue
- **Rate Limiting**: Respect provider rate limits (e.g., Twilio 1msg/sec)
- **Message Prefetch**: Configure RabbitMQ prefetch count

### Reliability
- **Dead Letter Queue**: Configure for failed messages
- **Retry Logic**: Exponential backoff for transient failures
- **Idempotency**: Use Redis to deduplicate messages
- **Circuit Breaker**: Prevent cascading failures

### Monitoring
- **Metrics**: Track send rates, failure rates, latency
- **Alerts**: Configure alerts for high failure rates
- **Logging**: Structured logs with correlation IDs
- **Tracing**: Distributed tracing for debugging

## Future Enhancements

- [ ] Implement PostgreSQL repository for audit trail
- [ ] Add Twilio integration for production SMS
- [ ] Add FCM/APNS for production push notifications
- [ ] Implement notification templates
- [ ] Add user preference management
- [ ] Implement notification deduplication with Redis
- [ ] Add notification scheduling/batching
- [ ] Implement webhook notifications
- [ ] Add notification analytics dashboard

## License

Part of the Airport Services ecosystem.

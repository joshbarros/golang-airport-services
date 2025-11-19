# Airport Services - Comprehensive Microservices Platform

A modern, scalable microservices architecture for comprehensive airport operations management, built with Go.

## Overview

This platform provides end-to-end digital solutions for airport operations, including:

- **Flight Operations**: Flight scheduling, aircraft management, crew assignments, gate management
- **Passenger Services**: Booking, check-in, boarding, passenger management
- **Baggage Operations**: Baggage tracking, lost & found, baggage handling
- **Commercial Services**: Payments, loyalty programs, ancillary services
- **Airport Operations**: Security, immigration, ground transportation, terminal management
- **Support Services**: Notifications, analytics, third-party integrations

## Architecture

This is a **microservices-based architecture** with 24 independent services organized by domain:

- 4 Core Flight Operations services
- 4 Passenger Services
- 2 Baggage Operations services
- 3 Commercial Services
- 4 Airport Operations services
- 3 Support Services
- 4 Infrastructure Services

For detailed architecture, see [ARCHITECTURE.md](./ARCHITECTURE.md)

### Architectural Principles

This project strictly follows industry best practices:

- **🏗️ Clean Architecture**: Clear separation of concerns with dependency inversion
- **🎯 Domain-Driven Design (DDD)**: Rich domain models with bounded contexts
- **⚡ SOLID Principles**: Maintainable and extensible code
- **🧪 Test-Driven Development (TDD)**: Tests written before production code

📚 **Comprehensive Guides Available**:
- [Clean Architecture Guide](./docs/CLEAN_ARCHITECTURE.md) - Layers, dependency rules, examples
- [DDD Guide](./docs/DDD_GUIDE.md) - Entities, aggregates, value objects, domain events
- [SOLID Principles Guide](./docs/SOLID_PRINCIPLES.md) - Practical Go examples
- [TDD Guide](./docs/TDD_GUIDE.md) - Red-Green-Refactor workflow
- [Architecture Docs Overview](./docs/README.md) - Quick reference and patterns

**All services must follow these principles** to ensure consistency and maintainability across the platform.

## Technology Stack

### Core Technologies
- **Language**: Go 1.21+
- **API**: REST (Gin) + gRPC
- **Databases**: PostgreSQL, MongoDB, Redis
- **Message Queue**: RabbitMQ / Kafka
- **Container**: Docker + Kubernetes
- **Observability**: Prometheus, Grafana, Jaeger

### Key Libraries
- Web: `gin-gonic/gin`, `grpc/grpc-go`
- Database: `jackc/pgx`, `mongodb/mongo-go-driver`
- Cache: `go-redis/redis`
- Logging: `uber-go/zap`
- Testing: `stretchr/testify`

## Project Structure

```
golang-airport-services/
├── cmd/                    # Service entry points
├── internal/               # Private service code
├── pkg/                    # Shared libraries
├── api/                    # API definitions (proto, openapi)
├── deployments/            # Docker, K8s, Terraform
├── migrations/             # Database migrations
├── test/                   # Integration, E2E, load tests
├── docs/                   # Documentation
└── scripts/                # Build and utility scripts
```

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- kubectl (for Kubernetes deployment)
- Make

### Local Development Setup

1. **Clone the repository**
```bash
git clone https://github.com/joshbarros/golang-airport-services.git
cd golang-airport-services
```

2. **Install dependencies**
```bash
make deps
```

3. **Start infrastructure services**
```bash
docker-compose up -d
```

4. **Run database migrations**
```bash
make migrate-up
```

5. **Start a service (example: Flight Service)**
```bash
make run-flight-service
```

### Build All Services

```bash
make build-all
```

### Run Tests

```bash
# Unit tests
make test

# Integration tests
make test-integration

# All tests with coverage
make test-coverage
```

## Development Phases

### ✅ Phase 1: Foundation (Completed)
- Project structure
- Architecture documentation

### 🚧 Phase 2: Infrastructure Setup (In Progress)
- CI/CD pipelines
- Kubernetes setup
- Development environment
- Core infrastructure services

### 📋 Phase 3: Core Services (Planned)
- Flight Service
- Booking Service
- Passenger Service
- Payment Service

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the complete roadmap.

## Services

| Service | Port | Status | Description |
|---------|------|--------|-------------|
| API Gateway | 8080 | 🚧 | Main entry point |
| Flight Service | 8081 | 📋 | Flight management |
| Booking Service | 8082 | 📋 | Reservations |
| Passenger Service | 8083 | 📋 | Passenger data |
| Payment Service | 8084 | 📋 | Payment processing |
| Check-in Service | 8085 | 📋 | Check-in operations |
| Baggage Service | 8086 | 📋 | Baggage tracking |
| Notification Service | 8087 | 📋 | Multi-channel notifications |
| Auth Service | 8088 | 🚧 | Authentication & authorization |

Legend: ✅ Complete | 🚧 In Progress | 📋 Planned

## API Documentation

Once services are running, API documentation is available at:

- **Swagger UI**: http://localhost:8080/swagger
- **GraphQL Playground**: http://localhost:8080/graphql
- **gRPC Reflection**: Use tools like grpcurl or BloomRPC

## Configuration

Services are configured via:
1. Environment variables
2. Configuration files (`config/`)
3. Kubernetes ConfigMaps/Secrets (production)

Example configuration:
```yaml
server:
  port: 8081
  host: 0.0.0.0

database:
  host: postgres
  port: 5432
  database: flight_db
  user: ${DB_USER}
  password: ${DB_PASSWORD}

redis:
  host: redis
  port: 6379

logger:
  level: info
  format: json
```

## Monitoring

### Local Development
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger**: http://localhost:16686

### Health Checks
Each service exposes:
- `GET /health` - Health status
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

## Database Migrations

```bash
# Create new migration
make migrate-create name=add_flights_table service=flight-service

# Apply migrations
make migrate-up service=flight-service

# Rollback migrations
make migrate-down service=flight-service

# Migration status
make migrate-status service=flight-service
```

## Testing Strategy

### Unit Tests
- Test individual functions and methods
- Mock external dependencies
- Target: 80% code coverage

```bash
make test-unit
```

### Integration Tests
- Test service interactions
- Use testcontainers for dependencies
- Test database operations

```bash
make test-integration
```

### E2E Tests
- Test complete workflows
- Simulate real user scenarios
- Run against staging environment

```bash
make test-e2e
```

### Load Tests
- Performance testing with k6
- Stress testing
- Capacity planning

```bash
make test-load
```

## Deployment

### Docker Compose (Development)
```bash
docker-compose up -d
```

### Kubernetes (Production)

```bash
# Deploy to development
kubectl apply -k deployments/kubernetes/overlays/dev

# Deploy to staging
kubectl apply -k deployments/kubernetes/overlays/staging

# Deploy to production
kubectl apply -k deployments/kubernetes/overlays/prod
```

### Infrastructure as Code (Terraform)

```bash
cd deployments/terraform
terraform init
terraform plan
terraform apply
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Standards
- Follow Go best practices and idioms
- Use `gofmt` and `goimports` for formatting
- Write unit tests for new code
- Update documentation as needed
- Add comments for complex logic

## Security

### Reporting Vulnerabilities
Please report security vulnerabilities to security@example.com

### Security Measures
- Authentication via JWT tokens
- Role-based access control (RBAC)
- Encryption at rest and in transit
- Regular security audits
- Dependency vulnerability scanning

## Performance

### Target Metrics
- API Response Time: p95 < 200ms, p99 < 500ms
- Throughput: 10,000 requests/second
- Availability: 99.99% uptime
- Database Query Time: p95 < 50ms

### Optimization Strategies
- Redis caching for hot data
- Database query optimization
- Connection pooling
- Horizontal scaling with Kubernetes HPA
- CDN for static assets

## Monitoring & Alerts

### Key Metrics
- Request rate and latency
- Error rates
- Resource utilization (CPU, memory)
- Database performance
- Message queue depth

### Alerting
- Critical: Service down, high error rate
- Warning: High latency, resource pressure
- Info: Deployment events, scaling events

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs/](./docs/)
- **Issue Tracker**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Email**: support@example.com

## Acknowledgments

- IATA standards and best practices
- Go community and open-source projects
- Microservices patterns and practices

---

**Status**: 🚧 In Active Development
**Version**: 0.1.0-alpha
**Last Updated**: 2025-11-19

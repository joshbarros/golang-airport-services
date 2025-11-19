# Contributing to Airport Services

Thank you for your interest in contributing to the Airport Services project! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors, regardless of background or identity.

### Expected Behavior

- Be respectful and considerate
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept responsibility for mistakes
- Show empathy towards other community members

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Trolling or insulting/derogatory comments
- Public or private harassment
- Publishing others' private information
- Other conduct inappropriate for a professional setting

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Git
- Make
- A code editor (VS Code, GoLand, etc.)

### Setup Development Environment

1. **Fork and clone the repository**
```bash
git clone https://github.com/YOUR_USERNAME/golang-airport-services.git
cd golang-airport-services
```

2. **Add upstream remote**
```bash
git remote add upstream https://github.com/joshbarros/golang-airport-services.git
```

3. **Install dependencies**
```bash
make deps
make tools
```

4. **Start infrastructure services**
```bash
make docker-up
```

5. **Run tests to verify setup**
```bash
make test
```

## Development Workflow

### 1. Create a Branch

Always create a new branch for your work:

```bash
# Update your fork
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/bug-description
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates
- `test/` - Test improvements
- `chore/` - Maintenance tasks

### 2. Make Changes Using TDD

We follow **Test-Driven Development (TDD)**. See [docs/TDD_GUIDE.md](./docs/TDD_GUIDE.md) for details.

**TDD Cycle**:
1. 🔴 **RED**: Write a failing test first
2. 🟢 **GREEN**: Write minimal code to make it pass
3. 🔵 **REFACTOR**: Improve code quality
4. 🔁 **REPEAT**: Continue with next test

**Required**:
- ✅ Write tests BEFORE production code
- ✅ Follow Clean Architecture principles
- ✅ Apply DDD patterns (Entities, Value Objects, Aggregates)
- ✅ Adhere to SOLID principles
- ✅ Add comprehensive tests (unit, integration)
- ✅ Update documentation as needed
- ✅ Keep commits atomic and focused

**Architecture Guides**:
- [Clean Architecture](./docs/CLEAN_ARCHITECTURE.md)
- [Domain-Driven Design](./docs/DDD_GUIDE.md)
- [SOLID Principles](./docs/SOLID_PRINCIPLES.md)
- [TDD Workflow](./docs/TDD_GUIDE.md)

### 3. Test Your Changes

```bash
# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run linter
make lint

# Check code coverage
make test-coverage
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "type: brief description"
```

See [Commit Message Guidelines](#commit-message-guidelines) for details.

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Coding Standards

### Go Style Guide

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html).

### Key Principles

1. **Simplicity**: Write simple, clear code
2. **Readability**: Code is read more than written
3. **Consistency**: Follow existing patterns
4. **Documentation**: Document exported functions and types
5. **Error Handling**: Always handle errors explicitly

### Code Formatting

```bash
# Format code
make fmt

# Run linter
make lint

# Auto-fix linter issues
make lint-fix
```

### Naming Conventions

**Packages**
```go
// Good
package flight
package booking

// Bad
package flightService
package Booking
```

**Functions**
```go
// Good
func CreateBooking() {}
func getInternalData() {}

// Bad
func create_booking() {}
func GetInternalData() {} // if not exported
```

**Variables**
```go
// Good
var userID string
var flightNumber int

// Bad
var user_id string
var FlightNumber int // unless constant
```

**Interfaces**
```go
// Good
type Reader interface {}
type FlightRepository interface {}

// Bad
type IReader interface {}
type FlightRepositoryInterface interface {}
```

### Error Handling

```go
// Good
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Bad
result, _ := doSomething()
```

### Package Structure

```go
// internal/flight/
//   domain/      - Domain models and business logic
//   handler/     - HTTP/gRPC handlers
//   repository/  - Data access layer
//   service/     - Business logic orchestration
//   event/       - Event publishers/subscribers
```

## Testing Guidelines

### Test Coverage

- Aim for 80%+ code coverage
- Focus on business logic and critical paths
- Don't test generated code or simple getters/setters

### Test Organization

```go
// flight_service_test.go
package flight_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateFlight(t *testing.T) {
    // Arrange
    service := NewFlightService()

    // Act
    result, err := service.CreateFlight(/* params */)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Test Types

**Unit Tests**
```go
// Test individual functions
func TestValidateFlightNumber(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "AA123", false},
        {"invalid", "123", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateFlightNumber(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Integration Tests**
```go
// +build integration

func TestFlightRepository(t *testing.T) {
    // Use testcontainers for database
    db := setupTestDatabase(t)
    defer db.Close()

    repo := NewFlightRepository(db)
    // Test database operations
}
```

**Mocking**
```go
// Use interfaces for dependencies
type FlightRepository interface {
    Create(flight *Flight) error
    FindByID(id string) (*Flight, error)
}

// Mock in tests
type MockFlightRepository struct {
    mock.Mock
}

func (m *MockFlightRepository) Create(flight *Flight) error {
    args := m.Called(flight)
    return args.Error(0)
}
```

## Commit Message Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks
- **perf**: Performance improvements
- **ci**: CI/CD changes

### Scope

The scope specifies the service or component:
- `flight`
- `booking`
- `passenger`
- `payment`
- `api-gateway`
- `common`

### Examples

```bash
feat(flight): add flight status update endpoint

Implement PUT /api/v1/flights/{id}/status endpoint to allow
real-time flight status updates.

Closes #123
```

```bash
fix(booking): prevent duplicate booking creation

Add unique constraint on booking table to prevent duplicate
bookings when user clicks submit multiple times.

Fixes #456
```

```bash
docs(readme): update setup instructions

Add instructions for Apple Silicon Mac users
```

## Pull Request Process

### Before Submitting

- [ ] Code follows project style guidelines
- [ ] All tests pass (`make test`)
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] Commits follow commit message guidelines
- [ ] Branch is up to date with main

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe testing performed

## Checklist
- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### Review Process

1. **Automated Checks**: CI/CD runs tests and linters
2. **Code Review**: At least one maintainer reviews
3. **Discussion**: Address feedback and make changes
4. **Approval**: Maintainer approves PR
5. **Merge**: Squash and merge into main

### Getting Your PR Merged

- Respond to feedback promptly
- Keep PRs focused and reasonably sized
- Be patient and respectful
- Update your branch if requested

## Project Structure

```
golang-airport-services/
├── cmd/                    # Application entry points
│   └── service-name/
│       └── main.go
├── internal/               # Private application code
│   └── service-name/
│       ├── domain/        # Business logic and models
│       ├── handler/       # HTTP/gRPC handlers
│       ├── repository/    # Data access
│       ├── service/       # Service layer
│       └── event/         # Event handling
├── pkg/                    # Shared libraries
│   ├── logger/
│   ├── database/
│   └── middleware/
├── api/                    # API definitions
│   ├── proto/
│   └── openapi/
├── deployments/            # Deployment configs
├── migrations/             # Database migrations
└── test/                   # Tests
```

### Adding a New Service

1. Create service directory structure
2. Implement domain models
3. Add repository layer
4. Implement service layer
5. Add HTTP/gRPC handlers
6. Write tests
7. Add database migrations
8. Update API documentation
9. Add deployment configs
10. Update main README

## Questions?

- Check existing issues and discussions
- Ask in GitHub Discussions
- Contact maintainers

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.

---

Thank you for contributing to Airport Services! 🚀✈️

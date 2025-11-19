package event

import "context"

// Publisher defines the interface for publishing domain events
// This follows the Dependency Inversion Principle - domain defines the interface,
// infrastructure provides the implementation (RabbitMQ, Kafka, etc.)
type Publisher interface {
	// Publish publishes a domain event
	Publish(ctx context.Context, event DomainEvent) error

	// PublishBatch publishes multiple domain events
	PublishBatch(ctx context.Context, events []DomainEvent) error

	// Close closes the publisher connection
	Close() error
}

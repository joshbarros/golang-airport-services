package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/domain/event"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// EventPublisher implements event.Publisher using RabbitMQ
type EventPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	logger   *logger.Logger
}

// NewEventPublisher creates a new RabbitMQ event publisher
func NewEventPublisher(url, exchange string, log *logger.Logger) (*EventPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange (topic exchange for event routing)
	err = channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Info("Connected to RabbitMQ",
		zap.String("exchange", exchange),
	)

	return &EventPublisher{
		conn:     conn,
		channel:  channel,
		exchange: exchange,
		logger:   log,
	}, nil
}

// Publish publishes a domain event to RabbitMQ
func (p *EventPublisher) Publish(ctx context.Context, evt event.DomainEvent) error {
	// Marshal event to JSON
	body, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create message
	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // Persistent messages
		Timestamp:    evt.OccurredAt(),
		Type:         evt.EventType(),
		MessageId:    fmt.Sprintf("%s-%d", evt.EventID(), time.Now().UnixNano()),
		Body:         body,
	}

	// Publish with routing key = event type (e.g., "flight.created")
	err = p.channel.PublishWithContext(
		ctx,
		p.exchange,       // exchange
		evt.EventType(),  // routing key
		false,            // mandatory
		false,            // immediate
		msg,
	)

	if err != nil {
		p.logger.Error("Failed to publish event",
			zap.String("event_type", evt.EventType()),
			zap.String("event_id", evt.EventID()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Debug("Published event",
		zap.String("event_type", evt.EventType()),
		zap.String("event_id", evt.EventID()),
	)

	return nil
}

// PublishBatch publishes multiple events (not implemented for now, would use transactions)
func (p *EventPublisher) PublishBatch(ctx context.Context, events []event.DomainEvent) error {
	for _, evt := range events {
		if err := p.Publish(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the RabbitMQ connection
func (p *EventPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			p.logger.Error("Failed to close channel", zap.Error(err))
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			p.logger.Error("Failed to close connection", zap.Error(err))
			return err
		}
	}
	p.logger.Info("Closed RabbitMQ connection")
	return nil
}

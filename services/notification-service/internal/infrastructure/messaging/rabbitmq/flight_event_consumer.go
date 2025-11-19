package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/usecase"
	"go.uber.org/zap"
)

// FlightEventConsumer consumes flight events from RabbitMQ and triggers notifications
type FlightEventConsumer struct {
	conn                     *amqp.Connection
	channel                  *amqp.Channel
	exchange                 string
	queue                    string
	sendNotificationUseCase  *usecase.SendNotificationUseCase
	logger                   *logger.Logger
	done                     chan bool
}

// FlightEvent represents a generic flight event structure
type FlightEvent struct {
	EventType       string                 `json:"event_type"`
	FlightID        string                 `json:"flight_id"`
	FlightNumber    string                 `json:"flight_number"`
	Origin          string                 `json:"origin"`
	Destination     string                 `json:"destination"`
	DelayDuration   int                    `json:"delay_duration"`
	Reason          string                 `json:"reason"`
	NewDeparture    time.Time              `json:"new_departure_time"`
	NewArrival      time.Time              `json:"new_arrival_time"`
	Gate            string                 `json:"gate"`
	Terminal        string                 `json:"terminal"`
	Metadata        map[string]interface{} `json:"-"`
}

// NewFlightEventConsumer creates a new event consumer
func NewFlightEventConsumer(
	url, exchange string,
	sendNotificationUC *usecase.SendNotificationUseCase,
	log *logger.Logger,
) (*FlightEventConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange (should already exist from Flight Service)
	err = channel.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue for notification service
	queueName := "notification-service-queue"
	queue, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange for all flight events
	eventPatterns := []string{
		"flight.created",
		"flight.status.changed",
		"flight.delayed",
		"flight.cancelled",
		"flight.boarding.started",
		"flight.departed",
		"flight.arrived",
	}

	for _, pattern := range eventPatterns {
		err = channel.QueueBind(
			queue.Name,
			pattern,  // routing key
			exchange,
			false,
			nil,
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue to pattern %s: %w", pattern, err)
		}
	}

	log.Info("Flight event consumer initialized",
		zap.String("exchange", exchange),
		zap.String("queue", queue.Name),
		zap.Strings("patterns", eventPatterns),
	)

	return &FlightEventConsumer{
		conn:                     conn,
		channel:                  channel,
		exchange:                 exchange,
		queue:                    queue.Name,
		sendNotificationUseCase:  sendNotificationUC,
		logger:                   log,
		done:                     make(chan bool),
	}, nil
}

// Start begins consuming messages
func (c *FlightEventConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"notification-service", // consumer tag
		false,                  // auto-ack (we'll manually ack)
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info("Started consuming flight events")

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Warn("Message channel closed")
					return
				}

				c.handleMessage(ctx, msg)

			case <-c.done:
				c.logger.Info("Consumer stopped")
				return

			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping consumer")
				return
			}
		}
	}()

	return nil
}

// handleMessage processes a single message
func (c *FlightEventConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	eventType := msg.Type

	c.logger.Info("Received flight event",
		zap.String("event_type", eventType),
		zap.String("routing_key", msg.RoutingKey),
	)

	// Parse event
	var event FlightEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Error("Failed to parse event",
			zap.Error(err),
			zap.String("event_type", eventType),
		)
		msg.Nack(false, false) // Don't requeue malformed messages
		return
	}

	event.EventType = eventType

	// Generate notification based on event type
	notifications := c.generateNotifications(event)

	// Send each notification
	successCount := 0
	for _, notif := range notifications {
		_, err := c.sendNotificationUseCase.Execute(ctx, notif)
		if err != nil {
			c.logger.Error("Failed to send notification",
				zap.Error(err),
				zap.String("event_type", eventType),
				zap.String("recipient", notif.Recipient),
			)
		} else {
			successCount++
		}
	}

	// Acknowledge message if at least one notification succeeded or no notifications were needed
	if successCount > 0 || len(notifications) == 0 {
		msg.Ack(false)
		c.logger.Info("Processed flight event",
			zap.String("event_type", eventType),
			zap.Int("notifications_sent", successCount),
		)
	} else {
		// Requeue for retry
		msg.Nack(false, true)
		c.logger.Warn("Requeuing message due to notification failures",
			zap.String("event_type", eventType),
		)
	}
}

// generateNotifications creates notification inputs based on flight event
func (c *FlightEventConsumer) generateNotifications(event FlightEvent) []usecase.SendNotificationInput {
	notifications := []usecase.SendNotificationInput{}

	// In a real system, you would:
	// 1. Query Passenger Service to get passengers on the flight
	// 2. Get passenger notification preferences
	// 3. Generate personalized notifications

	// For demo purposes, we'll create a sample notification
	// In production, this would come from passenger data
	var subject, message string

	switch event.EventType {
	case "flight.created":
		subject = fmt.Sprintf("Flight %s Scheduled", event.FlightNumber)
		message = fmt.Sprintf("Flight %s from %s to %s has been scheduled.",
			event.FlightNumber, event.Origin, event.Destination)

	case "flight.delayed":
		subject = fmt.Sprintf("Flight %s Delayed", event.FlightNumber)
		message = fmt.Sprintf("Flight %s has been delayed by %d minutes. Reason: %s. New departure time: %s",
			event.FlightNumber, event.DelayDuration, event.Reason, event.NewDeparture.Format("15:04"))

	case "flight.cancelled":
		subject = fmt.Sprintf("Flight %s Cancelled", event.FlightNumber)
		message = fmt.Sprintf("Flight %s from %s to %s has been cancelled. Reason: %s",
			event.FlightNumber, event.Origin, event.Destination, event.Reason)

	case "flight.boarding.started":
		subject = fmt.Sprintf("Boarding Started - Flight %s", event.FlightNumber)
		message = fmt.Sprintf("Boarding has started for flight %s at gate %s, terminal %s. Please proceed to the gate.",
			event.FlightNumber, event.Gate, event.Terminal)

	case "flight.departed":
		subject = fmt.Sprintf("Flight %s Departed", event.FlightNumber)
		message = fmt.Sprintf("Flight %s has departed from %s to %s.",
			event.FlightNumber, event.Origin, event.Destination)

	case "flight.arrived":
		subject = fmt.Sprintf("Flight %s Arrived", event.FlightNumber)
		message = fmt.Sprintf("Flight %s has arrived at %s.",
			event.FlightNumber, event.Destination)

	default:
		c.logger.Warn("Unknown event type", zap.String("event_type", event.EventType))
		return notifications
	}

	// Demo notification (in production, this would be per passenger)
	// For critical events (delays, cancellations), send to demo email
	if event.EventType == "flight.delayed" || event.EventType == "flight.cancelled" {
		notifications = append(notifications, usecase.SendNotificationInput{
			Type:         "email",
			Recipient:    "passenger@example.com", // Demo recipient
			Subject:      subject,
			Message:      message,
			FlightID:     event.FlightID,
			FlightNumber: event.FlightNumber,
			EventType:    event.EventType,
			Metadata:     map[string]interface{}{
				"origin":      event.Origin,
				"destination": event.Destination,
			},
		})
	}

	return notifications
}

// Stop stops the consumer
func (c *FlightEventConsumer) Stop() error {
	close(c.done)

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	c.logger.Info("Flight event consumer stopped")
	return nil
}

package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/valueobject"
)

// Notification is the aggregate root for notification management
type Notification struct {
	id             string
	notificationType valueobject.NotificationType
	status         valueobject.NotificationStatus
	recipient      string // email, phone number, or device token
	subject        string // for email
	message        string
	flightID       string // optional: associated flight
	flightNumber   string // optional: for display
	eventType      string // flight.created, flight.delayed, etc.
	metadata       map[string]interface{}
	sentAt         *time.Time
	failureReason  string
	retryCount     int
	maxRetries     int
	createdAt      time.Time
	updatedAt      time.Time
}

// NewNotification creates a new notification
func NewNotification(
	notificationType valueobject.NotificationType,
	recipient, subject, message string,
	flightID, flightNumber, eventType string,
	metadata map[string]interface{},
) (*Notification, error) {
	if recipient == "" {
		return nil, errors.BadRequest("recipient is required")
	}

	if message == "" {
		return nil, errors.BadRequest("message is required")
	}

	// Email requires subject
	if notificationType.IsEmail() && subject == "" {
		return nil, errors.BadRequest("subject is required for email notifications")
	}

	now := time.Now()
	return &Notification{
		id:               uuid.New().String(),
		notificationType: notificationType,
		status:           valueobject.NotificationStatusPending,
		recipient:        recipient,
		subject:          subject,
		message:          message,
		flightID:         flightID,
		flightNumber:     flightNumber,
		eventType:        eventType,
		metadata:         metadata,
		retryCount:       0,
		maxRetries:       3,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

// MarkAsSending marks notification as currently being sent
func (n *Notification) MarkAsSending() error {
	newStatus := valueobject.NotificationStatusSending
	if !n.status.CanTransitionTo(newStatus) {
		return errors.BadRequest("cannot mark as sending from current status")
	}
	n.status = newStatus
	n.updatedAt = time.Now()
	return nil
}

// MarkAsSent marks notification as successfully sent
func (n *Notification) MarkAsSent() error {
	newStatus := valueobject.NotificationStatusSent
	if !n.status.CanTransitionTo(newStatus) {
		return errors.BadRequest("cannot mark as sent from current status")
	}
	n.status = newStatus
	now := time.Now()
	n.sentAt = &now
	n.updatedAt = now
	return nil
}

// MarkAsFailed marks notification as failed with reason
func (n *Notification) MarkAsFailed(reason string) error {
	newStatus := valueobject.NotificationStatusFailed
	if !n.status.CanTransitionTo(newStatus) {
		return errors.BadRequest("cannot mark as failed from current status")
	}
	n.status = newStatus
	n.failureReason = reason
	n.updatedAt = time.Now()
	return nil
}

// Retry attempts to retry a failed notification
func (n *Notification) Retry() error {
	if n.retryCount >= n.maxRetries {
		return errors.BadRequest("maximum retry attempts exceeded")
	}

	newStatus := valueobject.NotificationStatusRetrying
	if !n.status.CanTransitionTo(newStatus) {
		return errors.BadRequest("cannot retry from current status")
	}

	n.status = newStatus
	n.retryCount++
	n.updatedAt = time.Now()
	return nil
}

// Cancel cancels a pending notification
func (n *Notification) Cancel() error {
	if n.status.IsFinal() {
		return errors.BadRequest("cannot cancel notification in final status")
	}

	newStatus := valueobject.NotificationStatusCancelled
	if !n.status.CanTransitionTo(newStatus) {
		return errors.BadRequest("cannot cancel from current status")
	}

	n.status = newStatus
	n.updatedAt = time.Now()
	return nil
}

// Getters
func (n *Notification) ID() string                                { return n.id }
func (n *Notification) Type() valueobject.NotificationType        { return n.notificationType }
func (n *Notification) Status() valueobject.NotificationStatus    { return n.status }
func (n *Notification) Recipient() string                         { return n.recipient }
func (n *Notification) Subject() string                           { return n.subject }
func (n *Notification) Message() string                           { return n.message }
func (n *Notification) FlightID() string                          { return n.flightID }
func (n *Notification) FlightNumber() string                      { return n.flightNumber }
func (n *Notification) EventType() string                         { return n.eventType }
func (n *Notification) Metadata() map[string]interface{}          { return n.metadata }
func (n *Notification) SentAt() *time.Time                        { return n.sentAt }
func (n *Notification) FailureReason() string                     { return n.failureReason }
func (n *Notification) RetryCount() int                           { return n.retryCount }
func (n *Notification) MaxRetries() int                           { return n.maxRetries }
func (n *Notification) CreatedAt() time.Time                      { return n.createdAt }
func (n *Notification) UpdatedAt() time.Time                      { return n.updatedAt }

// Reconstruct recreates a notification from persistence (used by repository)
func Reconstruct(
	id string,
	notificationType valueobject.NotificationType,
	status valueobject.NotificationStatus,
	recipient, subject, message string,
	flightID, flightNumber, eventType string,
	metadata map[string]interface{},
	sentAt *time.Time,
	failureReason string,
	retryCount, maxRetries int,
	createdAt, updatedAt time.Time,
) *Notification {
	return &Notification{
		id:               id,
		notificationType: notificationType,
		status:           status,
		recipient:        recipient,
		subject:          subject,
		message:          message,
		flightID:         flightID,
		flightNumber:     flightNumber,
		eventType:        eventType,
		metadata:         metadata,
		sentAt:           sentAt,
		failureReason:    failureReason,
		retryCount:       retryCount,
		maxRetries:       maxRetries,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

package repository

import (
	"context"

	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/valueobject"
)

// NotificationProvider defines the interface for sending notifications
// Infrastructure layer will implement this for different channels (Email, SMS, Push)
type NotificationProvider interface {
	// Type returns the notification type this provider handles
	Type() valueobject.NotificationType

	// Send sends a notification and returns error if failed
	Send(ctx context.Context, recipient, subject, message string, metadata map[string]interface{}) error

	// IsAvailable checks if the provider is configured and ready
	IsAvailable() bool
}

package repository

import (
	"context"

	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/entity"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/valueobject"
)

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	// Save creates or updates a notification
	Save(ctx context.Context, notification *entity.Notification) error

	// FindByID retrieves a notification by ID
	FindByID(ctx context.Context, id string) (*entity.Notification, error)

	// FindByFlightID retrieves all notifications for a flight
	FindByFlightID(ctx context.Context, flightID string, limit, offset int) ([]*entity.Notification, error)

	// FindByRecipient retrieves all notifications for a recipient
	FindByRecipient(ctx context.Context, recipient string, limit, offset int) ([]*entity.Notification, error)

	// FindByStatus retrieves notifications by status
	FindByStatus(ctx context.Context, status valueobject.NotificationStatus, limit, offset int) ([]*entity.Notification, error)

	// FindPendingForRetry finds failed notifications that can be retried
	FindPendingForRetry(ctx context.Context, limit int) ([]*entity.Notification, error)

	// FindAll retrieves all notifications with pagination
	FindAll(ctx context.Context, limit, offset int) ([]*entity.Notification, error)

	// Count returns total number of notifications
	Count(ctx context.Context) (int64, error)

	// CountByStatus returns count by status
	CountByStatus(ctx context.Context, status valueobject.NotificationStatus) (int64, error)
}

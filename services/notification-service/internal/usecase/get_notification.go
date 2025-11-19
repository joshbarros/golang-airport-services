package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/entity"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/repository"
)

// GetNotificationOutput represents notification details
type GetNotificationOutput struct {
	ID            string
	Type          string
	Status        string
	Recipient     string
	Subject       string
	Message       string
	FlightID      string
	FlightNumber  string
	EventType     string
	SentAt        *time.Time
	FailureReason string
	RetryCount    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// GetNotificationUseCase handles retrieving a notification by ID
type GetNotificationUseCase struct {
	notificationRepo repository.NotificationRepository
}

// NewGetNotificationUseCase creates a new instance
func NewGetNotificationUseCase(notificationRepo repository.NotificationRepository) *GetNotificationUseCase {
	return &GetNotificationUseCase{
		notificationRepo: notificationRepo,
	}
}

// Execute retrieves a notification by ID
func (uc *GetNotificationUseCase) Execute(ctx context.Context, id string) (*GetNotificationOutput, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.BadRequest("notification ID is required")
	}

	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if notification == nil {
		return nil, errors.NotFound("notification not found")
	}

	return toNotificationOutput(notification), nil
}

// Helper function to convert entity to output
func toNotificationOutput(n *entity.Notification) *GetNotificationOutput {
	return &GetNotificationOutput{
		ID:            n.ID(),
		Type:          n.Type().String(),
		Status:        n.Status().String(),
		Recipient:     n.Recipient(),
		Subject:       n.Subject(),
		Message:       n.Message(),
		FlightID:      n.FlightID(),
		FlightNumber:  n.FlightNumber(),
		EventType:     n.EventType(),
		SentAt:        n.SentAt(),
		FailureReason: n.FailureReason(),
		RetryCount:    n.RetryCount(),
		CreatedAt:     n.CreatedAt(),
		UpdatedAt:     n.UpdatedAt(),
	}
}

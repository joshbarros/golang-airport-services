package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/entity"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/repository"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/valueobject"
)

// SendNotificationInput represents the input for sending a notification
type SendNotificationInput struct {
	Type         string
	Recipient    string
	Subject      string
	Message      string
	FlightID     string
	FlightNumber string
	EventType    string
	Metadata     map[string]interface{}
}

// SendNotificationOutput represents the output after sending a notification
type SendNotificationOutput struct {
	ID           string
	Type         string
	Status       string
	Recipient    string
	FlightNumber string
	CreatedAt    time.Time
}

// SendNotificationUseCase handles sending notifications through various channels
type SendNotificationUseCase struct {
	notificationRepo repository.NotificationRepository
	providers        map[valueobject.NotificationType]repository.NotificationProvider
}

// NewSendNotificationUseCase creates a new instance
func NewSendNotificationUseCase(
	notificationRepo repository.NotificationRepository,
	providers []repository.NotificationProvider,
) *SendNotificationUseCase {
	providerMap := make(map[valueobject.NotificationType]repository.NotificationProvider)
	for _, provider := range providers {
		providerMap[provider.Type()] = provider
	}

	return &SendNotificationUseCase{
		notificationRepo: notificationRepo,
		providers:        providerMap,
	}
}

// Execute sends a notification
func (uc *SendNotificationUseCase) Execute(ctx context.Context, input SendNotificationInput) (*SendNotificationOutput, error) {
	// Validate notification type
	notificationType, err := valueobject.NewNotificationType(input.Type)
	if err != nil {
		return nil, err
	}

	// Check if provider is available
	provider, exists := uc.providers[notificationType]
	if !exists || !provider.IsAvailable() {
		return nil, errors.InternalServerError(
			fmt.Sprintf("notification provider for %s is not available", input.Type),
		)
	}

	// Create notification entity
	notification, err := entity.NewNotification(
		notificationType,
		input.Recipient,
		input.Subject,
		input.Message,
		input.FlightID,
		input.FlightNumber,
		input.EventType,
		input.Metadata,
	)
	if err != nil {
		return nil, err
	}

	// Save to repository first (audit trail)
	if err := uc.notificationRepo.Save(ctx, notification); err != nil {
		return nil, errors.InternalServerError("failed to save notification").Wrap(err)
	}

	// Mark as sending
	if err := notification.MarkAsSending(); err != nil {
		return nil, err
	}
	if err := uc.notificationRepo.Save(ctx, notification); err != nil {
		return nil, errors.InternalServerError("failed to update notification status").Wrap(err)
	}

	// Send notification through provider
	err = provider.Send(ctx, input.Recipient, input.Subject, input.Message, input.Metadata)
	if err != nil {
		// Mark as failed
		notification.MarkAsFailed(err.Error())
		uc.notificationRepo.Save(ctx, notification)
		return nil, errors.InternalServerError("failed to send notification").Wrap(err)
	}

	// Mark as sent
	if err := notification.MarkAsSent(); err != nil {
		return nil, err
	}
	if err := uc.notificationRepo.Save(ctx, notification); err != nil {
		return nil, errors.InternalServerError("failed to update notification status").Wrap(err)
	}

	return &SendNotificationOutput{
		ID:           notification.ID(),
		Type:         notification.Type().String(),
		Status:       notification.Status().String(),
		Recipient:    notification.Recipient(),
		FlightNumber: notification.FlightNumber(),
		CreatedAt:    notification.CreatedAt(),
	}, nil
}

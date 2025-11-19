package sms

import (
	"context"
	"fmt"

	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/valueobject"
	"go.uber.org/zap"
)

// ConsoleProvider implements SMS notification by logging to console
// In production, this would integrate with Twilio, AWS SNS, or similar
type ConsoleProvider struct {
	logger     *logger.Logger
	configured bool
}

// NewConsoleProvider creates a new console SMS provider
func NewConsoleProvider(log *logger.Logger) *ConsoleProvider {
	log.Info("SMS Console provider initialized (demo mode - logs to console)")

	return &ConsoleProvider{
		logger:     log,
		configured: true, // Always available in demo mode
	}
}

// Type returns the notification type
func (p *ConsoleProvider) Type() valueobject.NotificationType {
	return valueobject.NotificationTypeSMS
}

// IsAvailable checks if provider is configured
func (p *ConsoleProvider) IsAvailable() bool {
	return p.configured
}

// Send sends an SMS notification (logs to console in demo mode)
func (p *ConsoleProvider) Send(ctx context.Context, recipient, subject, message string, metadata map[string]interface{}) error {
	p.logger.Info("📱 SMS Notification (Console Demo)",
		zap.String("to", recipient),
		zap.String("message", message),
		zap.Any("metadata", metadata),
	)

	// In production, this would be:
	// return twilioClient.Messages.Create(recipient, message)

	fmt.Printf("\n")
	fmt.Printf("════════════════════════════════════════\n")
	fmt.Printf("📱 SMS NOTIFICATION\n")
	fmt.Printf("════════════════════════════════════════\n")
	fmt.Printf("To: %s\n", recipient)
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("════════════════════════════════════════\n")
	fmt.Printf("\n")

	return nil
}

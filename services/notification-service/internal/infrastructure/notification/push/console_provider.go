package push

import (
	"context"
	"fmt"

	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/valueobject"
	"go.uber.org/zap"
)

// ConsoleProvider implements push notification by logging to console
// In production, this would integrate with FCM (Firebase), APNS (Apple), or similar
type ConsoleProvider struct {
	logger     *logger.Logger
	configured bool
}

// NewConsoleProvider creates a new console push provider
func NewConsoleProvider(log *logger.Logger) *ConsoleProvider {
	log.Info("Push Console provider initialized (demo mode - logs to console)")

	return &ConsoleProvider{
		logger:     log,
		configured: true, // Always available in demo mode
	}
}

// Type returns the notification type
func (p *ConsoleProvider) Type() valueobject.NotificationType {
	return valueobject.NotificationTypePush
}

// IsAvailable checks if provider is configured
func (p *ConsoleProvider) IsAvailable() bool {
	return p.configured
}

// Send sends a push notification (logs to console in demo mode)
func (p *ConsoleProvider) Send(ctx context.Context, recipient, subject, message string, metadata map[string]interface{}) error {
	p.logger.Info("🔔 Push Notification (Console Demo)",
		zap.String("device_token", recipient),
		zap.String("title", subject),
		zap.String("body", message),
		zap.Any("metadata", metadata),
	)

	// In production, this would be:
	// return fcmClient.Send(ctx, &messaging.Message{
	//     Token: recipient,
	//     Notification: &messaging.Notification{
	//         Title: subject,
	//         Body:  message,
	//     },
	// })

	fmt.Printf("\n")
	fmt.Printf("════════════════════════════════════════\n")
	fmt.Printf("🔔 PUSH NOTIFICATION\n")
	fmt.Printf("════════════════════════════════════════\n")
	fmt.Printf("Device: %s\n", recipient)
	fmt.Printf("Title: %s\n", subject)
	fmt.Printf("Body: %s\n", message)
	fmt.Printf("════════════════════════════════════════\n")
	fmt.Printf("\n")

	return nil
}

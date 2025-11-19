package valueobject

import "github.com/joshbarros/golang-airport-services/pkg/errors"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

// Valid notification types
var validNotificationTypes = map[NotificationType]bool{
	NotificationTypeEmail: true,
	NotificationTypeSMS:   true,
	NotificationTypePush:  true,
}

// NewNotificationType creates and validates a notification type
func NewNotificationType(t string) (NotificationType, error) {
	notifType := NotificationType(t)
	if !validNotificationTypes[notifType] {
		return "", errors.BadRequest("invalid notification type: must be email, sms, or push")
	}
	return notifType, nil
}

// String returns the string representation
func (n NotificationType) String() string {
	return string(n)
}

// IsEmail checks if notification type is email
func (n NotificationType) IsEmail() bool {
	return n == NotificationTypeEmail
}

// IsSMS checks if notification type is SMS
func (n NotificationType) IsSMS() bool {
	return n == NotificationTypeSMS
}

// IsPush checks if notification type is push
func (n NotificationType) IsPush() bool {
	return n == NotificationTypePush
}

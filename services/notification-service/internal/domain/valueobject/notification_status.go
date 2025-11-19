package valueobject

import "github.com/joshbarros/golang-airport-services/pkg/errors"

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSending   NotificationStatus = "sending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusRetrying  NotificationStatus = "retrying"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// Valid notification statuses
var validNotificationStatuses = map[NotificationStatus]bool{
	NotificationStatusPending:   true,
	NotificationStatusSending:   true,
	NotificationStatusSent:      true,
	NotificationStatusFailed:    true,
	NotificationStatusRetrying:  true,
	NotificationStatusCancelled: true,
}

// Status transitions map
var statusTransitions = map[NotificationStatus][]NotificationStatus{
	NotificationStatusPending: {
		NotificationStatusSending,
		NotificationStatusCancelled,
	},
	NotificationStatusSending: {
		NotificationStatusSent,
		NotificationStatusFailed,
	},
	NotificationStatusFailed: {
		NotificationStatusRetrying,
		NotificationStatusCancelled,
	},
	NotificationStatusRetrying: {
		NotificationStatusSending,
		NotificationStatusCancelled,
	},
	NotificationStatusSent:      {},
	NotificationStatusCancelled: {},
}

// NewNotificationStatus creates and validates a notification status
func NewNotificationStatus(s string) (NotificationStatus, error) {
	status := NotificationStatus(s)
	if !validNotificationStatuses[status] {
		return "", errors.BadRequest("invalid notification status")
	}
	return status, nil
}

// String returns the string representation
func (n NotificationStatus) String() string {
	return string(n)
}

// CanTransitionTo checks if status can transition to target status
func (n NotificationStatus) CanTransitionTo(target NotificationStatus) bool {
	allowedTransitions, exists := statusTransitions[n]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == target {
			return true
		}
	}
	return false
}

// IsFinal checks if this is a final status (no further transitions)
func (n NotificationStatus) IsFinal() bool {
	return n == NotificationStatusSent || n == NotificationStatusCancelled
}

// IsSuccess checks if status represents successful delivery
func (n NotificationStatus) IsSuccess() bool {
	return n == NotificationStatusSent
}

// IsFailure checks if status represents failure
func (n NotificationStatus) IsFailure() bool {
	return n == NotificationStatusFailed || n == NotificationStatusCancelled
}

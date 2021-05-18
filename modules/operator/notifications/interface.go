package notifications

// NotificationChannel is the interface that is responsible for notification sending mostly.
// Its main method is SendNotification that sends a notification message.
// sending parameters that are specific for each implementation is represented as a map[string]string
// For example: {"to": "example@example.org"} contains required options for email channel.
type NotificationChannel interface {
	SendNotification(opts map[string]string, head, body string) error
}

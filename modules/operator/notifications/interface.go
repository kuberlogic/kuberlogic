package notifications

import (
	"fmt"
	"github.com/kuberlogic/operator/modules/operator/notifications/smtp"
)

// NotificationChannel is the interface that is responsible for notification sending mostly.
// Its main method is SendNotification that sends a notification message.
// sending parameters that are specific for each implementation is represented as a map[string]string
// For example: {"to": "example@example.org"} contains required options for email channel.
type NotificationChannel interface {
	SendNotification(opts map[string]string, head, body string) error
}

func GetNotificationChannel(name string) (NotificationChannel, error) {
	var ch NotificationChannel
	var err error

	switch name {
	case EmailChannel:
		ch, err = smtp.NewSmtpChannel()
	default:
		err = fmt.Errorf("no %s notification channel found", name)
	}
	return ch, err
}

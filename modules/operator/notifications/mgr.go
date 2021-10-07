package notifications

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/operator/notifications/smtp"
)

type NotificationManager struct {
	emailEnabled  bool
	emailSettings cfg.EmailNotificationChannelConfig
}

func (m *NotificationManager) GetNotificationChannel(name string) (NotificationChannel, error) {
	var ch NotificationChannel
	var err error

	switch name {
	case EmailChannel:
		if !m.emailEnabled {
			return nil, fmt.Errorf("email notification channel is not enabled")
		}
		ch, err = smtp.NewSmtpChannel(
			m.emailSettings.Host, m.emailSettings.Port,
			m.emailSettings.TLS.Enabled, m.emailSettings.TLS.Insecure,
			m.emailSettings.Username, m.emailSettings.Password)
	default:
		err = fmt.Errorf("no %s notification channel found", name)
	}
	return ch, err
}

func NewWithConfig(config *cfg.Config) *NotificationManager {
	m := new(NotificationManager)
	if config.NotificationChannels.EmailEnabled {
		m.emailEnabled = true
		m.emailSettings = config.NotificationChannels.Email
	}
	return m
}

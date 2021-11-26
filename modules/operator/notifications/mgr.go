/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
			m.emailSettings.Host, m.emailSettings.Port, m.emailSettings.From,
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

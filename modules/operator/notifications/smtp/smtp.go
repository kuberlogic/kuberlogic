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

package smtp

import (
	"fmt"
	smtpLib "net/smtp"
)

type smtpChannel struct {
	host string
	port int
	tls  struct {
		insecure bool
		enabled  bool
	}
	username string
	password string

	sentEmail string
}

func (s *smtpChannel) SendNotification(options map[string]string, head, body string) error {
	to, found := options["to"]
	if !found {
		return fmt.Errorf("recipient email is not found")
	}
	subject := head

	msg := []byte(
		"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n")

	err := smtpLib.SendMail(s.addr(), s.auth(), s.sentEmail, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("%s: %v", "error sending email", err)
	}
	return nil
}

func (s *smtpChannel) addr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

func (s *smtpChannel) auth() smtpLib.Auth {
	if s.username == "" {
		return nil
	}
	return smtpLib.PlainAuth("", s.username, s.password, s.host)
}

func NewSmtpChannel(host string, port int, tlsEnabled, tlsInsecure bool, username, password string) (*smtpChannel, error) {
	c := &smtpChannel{
		host: host,
		port: port,
		tls: struct {
			insecure bool
			enabled  bool
		}{insecure: tlsInsecure, enabled: tlsEnabled},
		username: username,
		password: password,
	}
	return c, nil
}

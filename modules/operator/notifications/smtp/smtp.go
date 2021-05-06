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

func NewSmtpChannel() (*smtpChannel, error) {
	c := &smtpChannel{
		host: "mailservice.default",
		port: 25,
		tls: struct {
			insecure bool
			enabled  bool
		}{insecure: false, enabled: false},
		username: "",
		password: "",
	}
	return c, nil
}

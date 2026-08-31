package mcfeely

import (
	"context"
	"net/smtp"
)

type SmtpMcFeely struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSmtpMcFeely(host, port, username, password, from string) *SmtpMcFeely {
	return &SmtpMcFeely{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SmtpMcFeely) Send(ctx context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	msg := []byte("To: " + to + "\r\n" +
		"From: " + s.from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := s.host + ":" + s.port
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}

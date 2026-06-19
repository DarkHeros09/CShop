package mail

import (
	"crypto/tls"
	"fmt"

	"github.com/wneessen/go-mail"
)

const (
	smtpAuthAddress = "smtp.gmail.com"
	// go-mail handles the port separately or parses it; we'll use 587 as an int
	smtpPort = 587
)

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (sender *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	// Initialize the message
	m := mail.NewMsg()

	// Set From address formatted nicely with a display name
	if err := m.FromFormat(sender.name, sender.fromEmailAddress); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, content)

	// Set bulk recipient slices
	m.To(to...)
	m.Cc(cc...)
	m.Bcc(bcc...)

	// Attach files if any
	for _, f := range attachFiles {
		// AttachFile handles file reading and content-type detection internally
		m.AttachFile(f)
	}

	// Initialize the SMTP client with modern configuration options
	c, err := mail.NewClient(
		smtpAuthAddress,
		mail.WithPort(smtpPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(sender.fromEmailAddress),
		mail.WithPassword(sender.fromEmailPassword),
		// Tell go-mail to explicitly upgrade the plain connection to STARTTLS
		mail.WithTLSPolicy(mail.TLSMandatory),
		// Custom TLS config for InsecureSkipVerify
		mail.WithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         smtpAuthAddress,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize smtp client: %w", err)
	}

	// Dial the server and transmit the message
	if err := c.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

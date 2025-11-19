package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/notification-service/internal/domain/valueobject"
	"go.uber.org/zap"
)

// SMTPProvider implements email notification via SMTP
type SMTPProvider struct {
	host       string
	port       string
	username   string
	password   string
	from       string
	useTLS     bool
	logger     *logger.Logger
	configured bool
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(host, port, username, password, from string, useTLS bool, log *logger.Logger) *SMTPProvider {
	configured := host != "" && port != "" && from != ""

	if configured {
		log.Info("SMTP provider configured",
			zap.String("host", host),
			zap.String("port", port),
			zap.String("from", from),
			zap.Bool("tls", useTLS),
		)
	} else {
		log.Warn("SMTP provider not configured, email notifications will be disabled")
	}

	return &SMTPProvider{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		from:       from,
		useTLS:     useTLS,
		logger:     log,
		configured: configured,
	}
}

// Type returns the notification type
func (p *SMTPProvider) Type() valueobject.NotificationType {
	return valueobject.NotificationTypeEmail
}

// IsAvailable checks if provider is configured
func (p *SMTPProvider) IsAvailable() bool {
	return p.configured
}

// Send sends an email notification
func (p *SMTPProvider) Send(ctx context.Context, recipient, subject, message string, metadata map[string]interface{}) error {
	if !p.configured {
		return fmt.Errorf("SMTP provider not configured")
	}

	// Validate email format
	if !strings.Contains(recipient, "@") {
		return fmt.Errorf("invalid email address: %s", recipient)
	}

	// Build email message
	headers := make(map[string]string)
	headers["From"] = p.from
	headers["To"] = recipient
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""

	emailMessage := ""
	for key, value := range headers {
		emailMessage += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	emailMessage += "\r\n" + message

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", p.host, p.port)

	// Send email
	var err error
	if p.useTLS {
		err = p.sendWithTLS(addr, recipient, []byte(emailMessage))
	} else {
		err = p.sendPlain(addr, recipient, []byte(emailMessage))
	}

	if err != nil {
		p.logger.Error("Failed to send email",
			zap.Error(err),
			zap.String("recipient", recipient),
			zap.String("subject", subject),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	p.logger.Info("Email sent successfully",
		zap.String("recipient", recipient),
		zap.String("subject", subject),
	)

	return nil
}

func (p *SMTPProvider) sendPlain(addr, recipient string, message []byte) error {
	auth := smtp.PlainAuth("", p.username, p.password, p.host)
	return smtp.SendMail(addr, auth, p.from, []string{recipient}, message)
}

func (p *SMTPProvider) sendWithTLS(addr, recipient string, message []byte) error {
	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: p.host,
	}

	// Connect to SMTP server with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, p.host)
	if err != nil {
		return err
	}
	defer client.Quit()

	// Authenticate if credentials provided
	if p.username != "" && p.password != "" {
		auth := smtp.PlainAuth("", p.username, p.password, p.host)
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	// Set sender
	if err = client.Mail(p.from); err != nil {
		return err
	}

	// Set recipient
	if err = client.Rcpt(recipient); err != nil {
		return err
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return err
	}

	_, err = writer.Write(message)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}

package notifications

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

var emailConfig EmailConfig

// InitializeEmailConfig sets up the email configuration
func InitializeEmailConfig() {
	emailConfig = EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}

	// Log configuration status (without sensitive data)
	log.Printf("Email configuration initialized with host: %s, port: %s, from: %s",
		emailConfig.Host,
		emailConfig.Port,
		emailConfig.From)
}

// TestEmailConfig verifies if the email configuration is valid
func TestEmailConfig() error {
	if emailConfig.Host == "" {
		return fmt.Errorf("SMTP_HOST is not set")
	}
	if emailConfig.Port == "" {
		return fmt.Errorf("SMTP_PORT is not set")
	}
	if emailConfig.Username == "" {
		return fmt.Errorf("SMTP_USERNAME is not set")
	}
	if emailConfig.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD is not set")
	}
	if emailConfig.From == "" {
		return fmt.Errorf("SMTP_FROM is not set")
	}

	// Try to establish a connection to the SMTP server
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)
	addr := fmt.Sprintf("%s:%s", emailConfig.Host, emailConfig.Port)

	// Create a test connection
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer c.Close()

	// Try to authenticate
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate with SMTP server: %v", err)
	}

	log.Println("✅ Email configuration is valid and working!")
	return nil
}

// SendEmail sends an email to the specified recipient
func SendEmail(to, subject, body string) error {
	if emailConfig.Host == "" {
		return fmt.Errorf("email configuration not initialized")
	}

	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", emailConfig.From, to, subject, body)

	addr := fmt.Sprintf("%s:%s", emailConfig.Host, emailConfig.Port)
	err := smtp.SendMail(addr, auth, emailConfig.From, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}

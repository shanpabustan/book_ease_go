package notifications

import (
	"crypto/tls"
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
	log.Printf("📧 Email configuration initialized with host: %s, port: %s, from: %s",
		emailConfig.Host,
		emailConfig.Port,
		emailConfig.From)
}

// TestEmailConfig verifies if the email configuration is valid
func TestEmailConfig() error {
	log.Println("🔍 Testing email configuration...")

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

	log.Println("✅ Environment variables are set correctly")

	// Create auth
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)
	log.Println("✅ Auth created")

	// Create SMTP client
	addr := fmt.Sprintf("%s:%s", emailConfig.Host, emailConfig.Port)
	log.Printf("🔌 Connecting to SMTP server at %s...", addr)

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer client.Close()
	log.Println("✅ Connected to SMTP server")

	// Start TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         emailConfig.Host,
	}
	log.Println("🔒 Starting TLS...")

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %v", err)
	}
	log.Println("✅ TLS started successfully")

	// Authenticate
	log.Println("🔑 Authenticating...")
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate with SMTP server: %v", err)
	}
	log.Println("✅ Authentication successful")

	log.Println("✅ Email configuration is valid and working!")
	return nil
}

// SendEmail sends an email to the specified recipient
func SendEmail(to, subject, body string) error {
	log.Printf("📧 Attempting to send email to: %s", to)

	if emailConfig.Host == "" {
		return fmt.Errorf("email configuration not initialized")
	}

	// Create auth
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)
	log.Println("✅ Auth created")

	// Create SMTP client
	addr := fmt.Sprintf("%s:%s", emailConfig.Host, emailConfig.Port)
	log.Printf("🔌 Connecting to SMTP server at %s...", addr)

	client, err := smtp.Dial(addr)
	if err != nil {
		log.Printf("❌ Failed to connect to SMTP server: %v", err)
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer client.Close()
	log.Println("✅ Connected to SMTP server")

	// Start TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         emailConfig.Host,
	}
	log.Println("🔒 Starting TLS...")

	if err := client.StartTLS(tlsConfig); err != nil {
		log.Printf("❌ Failed to start TLS: %v", err)
		return fmt.Errorf("failed to start TLS: %v", err)
	}
	log.Println("✅ TLS started successfully")

	// Authenticate
	log.Println("🔑 Authenticating...")
	if err := client.Auth(auth); err != nil {
		log.Printf("❌ Failed to authenticate: %v", err)
		return fmt.Errorf("failed to authenticate with SMTP server: %v", err)
	}
	log.Println("✅ Authentication successful")

	// Set sender and recipient
	log.Printf("📨 Setting sender: %s", emailConfig.From)
	if err := client.Mail(emailConfig.From); err != nil {
		log.Printf("❌ Failed to set sender: %v", err)
		return fmt.Errorf("failed to set sender: %v", err)
	}

	log.Printf("📨 Setting recipient: %s", to)
	if err := client.Rcpt(to); err != nil {
		log.Printf("❌ Failed to set recipient: %v", err)
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send email data
	log.Println("📝 Preparing email data...")
	w, err := client.Data()
	if err != nil {
		log.Printf("❌ Failed to create email writer: %v", err)
		return fmt.Errorf("failed to create email writer: %v", err)
	}
	defer w.Close()

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", emailConfig.From, to, subject, body)

	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Printf("❌ Failed to write email data: %v", err)
		return fmt.Errorf("failed to write email data: %v", err)
	}
	log.Println("✅ Email data written successfully")

	log.Println("✅ Email sent successfully!")
	return nil
}

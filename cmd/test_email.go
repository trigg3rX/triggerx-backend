package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

func sendEmailNotification(to, subject, body string) error {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	emailFrom := os.Getenv("EMAIL_USER")
	emailPassword := os.Getenv("EMAIL_PASS")

	if emailFrom == "" || emailPassword == "" {
		return fmt.Errorf("EMAIL_USER and EMAIL_PASS environment variables must be set in .env file")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Create dialer with TLS
	d := gomail.NewDialer("smtp.zeptomail.in", 587, emailFrom, emailPassword)
	d.SSL = false
	d.TLSConfig = nil

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Email sent successfully to: %s", to)
	return nil
}

func main() {
	// Test email
	to := "nipunshah@lampros.tech"
	subject := "Test Email from TriggerX"
	body := `
		<h2>Test Email</h2>
		<p>This is a test email from TriggerX backend.</p>
		<p>If you're receiving this, the email notification system is working correctly.</p>
	`

	if err := sendEmailNotification(to, subject, body); err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}
	log.Println("Email sent successfully!")
}

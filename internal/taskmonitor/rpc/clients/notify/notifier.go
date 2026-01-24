package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"crypto/tls"
	"net"
	"net/smtp"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type Notifier interface {
	NotifyTaskStatus(ctx context.Context, email string, payload TaskStatusPayload) error
}

type WebhookNotifier struct {
	logger observability.Logger
	url    string
	token  string
	client *http.Client
}

type TaskStatusPayload struct {
	TaskID          int64     `json:"task_id"`
	JobID           int64     `json:"job_id"`
	Status          string    `json:"status"`
	IsAccepted      bool      `json:"is_accepted"`
	ExecutionTxHash string    `json:"execution_tx_hash,omitempty"`
	SubmissionTx    string    `json:"submission_tx_hash,omitempty"`
	ProofOfTask     string    `json:"proof_of_task,omitempty"`
	Error           string    `json:"error,omitempty"`
	OccurredAt      time.Time `json:"occurred_at"`
	Email           string    `json:"email"`
}

func NewWebhookNotifier(logger observability.Logger) *WebhookNotifier {
	return &WebhookNotifier{
		logger: logger.With(observability.String("component", "notifier")),
		url:    config.GetNotifyWebhookURL(),
		token:  config.GetNotifyWebhookToken(),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *WebhookNotifier) NotifyTaskStatus(ctx context.Context, email string, payload TaskStatusPayload) error {
	if n.url == "" {
		n.logger.Warn(ctx, "Notification webhook URL is not configured; skipping notification")
		return nil
	}
	payload.Email = email
	n.logger.Info(ctx, "Sending webhook notification", observability.String("email", email), observability.Int64("task_id", payload.TaskID), observability.String("status", payload.Status))
	body, err := json.Marshal(payload)
	if err != nil {
		n.logger.Error(ctx, "Failed to marshal webhook payload", observability.Error(err))
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader(body))
	if err != nil {
		n.logger.Error(ctx, "Failed to build webhook request", observability.Error(err))
		return fmt.Errorf("failed to build webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if n.token != "" {
		req.Header.Set("Authorization", "Bearer "+n.token)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		n.logger.Error(ctx, "Webhook request failed", observability.Error(err))
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			n.logger.Error(ctx, "Failed to close response body", observability.Error(closeErr))
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		n.logger.Error(ctx, "Webhook returned non-2xx status", observability.Int("status_code", resp.StatusCode))
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}
	n.logger.Info(ctx, "Webhook notification sent", observability.String("email", email), observability.Int64("task_id", payload.TaskID), observability.String("status", payload.Status))
	return nil
}

// SMTP notifier sends an email directly via SMTP
type SMTPNotifier struct {
	logger observability.Logger
}

func NewSMTPNotifier(logger observability.Logger) *SMTPNotifier {
	return &SMTPNotifier{logger: logger.With(observability.String("component", "smtp_notifier"))}
}

func (n *SMTPNotifier) NotifyTaskStatus(ctx context.Context, email string, payload TaskStatusPayload) error {
	host := config.GetSMTPHost()
	if host == "" {
		n.logger.Warn(ctx, "SMTP not configured; skipping email notification")
		return nil
	}
	port := config.GetSMTPPort()
	user := config.GetSMTPUser()
	pass := config.GetSMTPPass()
	from := config.GetSMTPFrom()
	useStartTLS := config.GetSMTPStartTLS()

	if from == "" {
		from = user
	}

	subject := "Task status update"
	if payload.Status == "failed" {
		subject = "Task failed notification"
	}
	body := fmt.Sprintf("Task %d (%s). Accepted: %t. SubmissionTx: %s, ExecutionTx: %s, Error: %s, At: %s",
		payload.TaskID, payload.Status, payload.IsAccepted, payload.SubmissionTx, payload.ExecutionTxHash, payload.Error, payload.OccurredAt.Format(time.RFC3339))

	// Build message
	msg := bytes.NewBuffer(nil)
	fmt.Fprintf(msg, "From: %s\r\n", from)
	fmt.Fprintf(msg, "To: %s\r\n", email)
	fmt.Fprintf(msg, "Subject: %s\r\n", subject)
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", host, port)
	n.logger.Info(ctx, "Attempting SMTP send", observability.String("host", host), observability.Int("port", port), observability.String("to", email), observability.Int64("task_id", payload.TaskID), observability.String("status", payload.Status))

	// Case 1: STARTTLS (587)
	if useStartTLS {
		d := &net.Dialer{Timeout: 10 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return fmt.Errorf("smtp dial failed: %w", err)
		}
		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp new client failed: %w", err)
		}
		defer func() {
			if quitErr := c.Quit(); quitErr != nil {
				n.logger.Error(ctx, "Failed to quit SMTP client", observability.Error(quitErr))
			}
		}()

		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: host}
			if err := c.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("smtp STARTTLS failed: %w", err)
			}
		}
		if user != "" && pass != "" {
			if err := c.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
				return fmt.Errorf("smtp auth failed: %w", err)
			}
		}
		if err := c.Mail(from); err != nil {
			return fmt.Errorf("smtp MAIL FROM failed: %w", err)
		}
		if err := c.Rcpt(email); err != nil {
			return fmt.Errorf("smtp RCPT TO failed: %w", err)
		}
		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("smtp DATA failed: %w", err)
		}
		if _, err := w.Write(msg.Bytes()); err != nil {
			return fmt.Errorf("smtp write body failed: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("smtp close writer failed: %w", err)
		}
		n.logger.Info(ctx, "SMTP email sent (STARTTLS)", observability.String("to", email))
		return nil
	}

	// Case 2: Implicit TLS (465)
	if port == 465 {
		tlsConfig := &tls.Config{ServerName: host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("smtps dial failed: %w", err)
		}
		defer func() {
			if closeErr := conn.Close(); closeErr != nil {
				n.logger.Error(ctx, "Failed to close connection", observability.Error(closeErr))
			}
		}()
		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtps new client failed: %w", err)
		}
		defer func() {
			if quitErr := c.Quit(); quitErr != nil {
				n.logger.Error(ctx, "Failed to quit SMTP client", observability.Error(quitErr))
			}
		}()

		if user != "" && pass != "" {
			if err := c.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
				return fmt.Errorf("smtps auth failed: %w", err)
			}
		}
		if err := c.Mail(from); err != nil {
			return fmt.Errorf("smtps MAIL FROM failed: %w", err)
		}
		if err := c.Rcpt(email); err != nil {
			return fmt.Errorf("smtps RCPT TO failed: %w", err)
		}
		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("smtps DATA failed: %w", err)
		}
		if _, err := w.Write(msg.Bytes()); err != nil {
			return fmt.Errorf("smtps write body failed: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("smtps close writer failed: %w", err)
		}
		n.logger.Info(ctx, "SMTP email sent (SMTPS)", observability.String("to", email))
		return nil
	}

	// Case 3: Plain (not recommended for Zoho, but fallback)
	auth := smtp.PlainAuth("", user, pass, host)
	if err := smtp.SendMail(addr, auth, from, []string{email}, msg.Bytes()); err != nil {
		return fmt.Errorf("smtp send mail failed: %w", err)
	}
	n.logger.Info(ctx, "SMTP email sent (plain)", observability.String("to", email))
	return nil
}

// Composite notifier calls multiple notifiers
type CompositeNotifier struct {
	notifiers []Notifier
	logger    observability.Logger
}

func NewCompositeNotifier(logger observability.Logger, notifiers ...Notifier) *CompositeNotifier {
	return &CompositeNotifier{notifiers: notifiers, logger: logger.With(observability.String("component", "composite_notifier"))}
}

func (c *CompositeNotifier) NotifyTaskStatus(ctx context.Context, email string, payload TaskStatusPayload) error {
	var firstErr error
	for _, n := range c.notifiers {
		if n == nil {
			continue
		}
		if err := n.NotifyTaskStatus(ctx, email, payload); err != nil {
			c.logger.Error(ctx, "Notifier failed", observability.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if firstErr == nil {
		c.logger.Info(ctx, "All notifiers processed", observability.String("email", email), observability.Int64("task_id", payload.TaskID), observability.String("status", payload.Status))
	}
	return firstErr
}

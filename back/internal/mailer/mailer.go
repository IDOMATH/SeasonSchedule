package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strings"

	"seasonschedule/internal/models"
)

// Config holds SMTP configuration loaded from environment variables.
type Config struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

// LoadConfig reads SMTP settings from environment variables.
// Defaults are provided for Gmail.
func LoadConfig() Config {
	return Config{
		Host: getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port: getEnv("SMTP_PORT", "587"),
		User: getEnv("SMTP_USER", ""),
		Pass: getEnv("SMTP_PASS", ""),
		From: getEnv("SMTP_FROM", ""),
	}
}

// Mailer handles sending emails.
type Mailer struct {
	cfg Config
}

// New creates a new Mailer with the provided config.
func New(cfg Config) *Mailer {
	return &Mailer{cfg: cfg}
}

// SendWeeklyDigest sends the weekly schedule digest to all subscriber emails.
func (m *Mailer) SendWeeklyDigest(emails []string, events []models.EventItem) error {
	if len(emails) == 0 {
		log.Println("Mailer: no subscribers, skipping digest")
		return nil
	}
	if m.cfg.User == "" || m.cfg.Pass == "" {
		log.Println("Mailer: SMTP credentials not configured, skipping digest")
		return nil
	}

	subject := "Your Weekly Schedule"
	body, err := buildEmailBody(events)
	if err != nil {
		return fmt.Errorf("mailer: failed to build email body: %w", err)
	}

	addr := m.cfg.Host + ":" + m.cfg.Port
	auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)

	from := m.cfg.From
	if from == "" {
		from = m.cfg.User
	}

	msg := buildMIMEMessage(from, subject, body)

	if err := smtp.SendMail(addr, auth, from, emails, msg); err != nil {
		return fmt.Errorf("mailer: smtp send failed: %w", err)
	}

	log.Printf("Mailer: weekly digest sent to %d subscriber(s)", len(emails))
	return nil
}

// buildMIMEMessage assembles a MIME email message.
func buildMIMEMessage(from, subject, htmlBody string) []byte {
	headers := strings.Join([]string{
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("Subject: %s", subject),
	}, "\r\n")

	return []byte(headers + "\r\n\r\n" + htmlBody)
}

var emailTmpl = template.Must(template.New("digest").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: sans-serif; background: #f4f4f4; margin: 0; padding: 0; }
    .container { max-width: 600px; margin: 2rem auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,.1); }
    .header { background: #667eea; color: #fff; padding: 1.5rem 2rem; }
    .header h1 { margin: 0; font-size: 1.4rem; }
    .header p { margin: 0.25rem 0 0; opacity: .85; font-size: .9rem; }
    .events { padding: 1.5rem 2rem; }
    .event { border-left: 4px solid #667eea; padding: .75rem 1rem; margin-bottom: 1rem; background: #f8f9ff; border-radius: 0 4px 4px 0; }
    .event h3 { margin: 0 0 .25rem; color: #2d3748; font-size: 1rem; }
    .event p  { margin: 0; color: #718096; font-size: .875rem; }
    .empty { color: #a0aec0; text-align: center; padding: 2rem; }
    .footer { text-align: center; padding: 1rem; font-size: .75rem; color: #a0aec0; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>📅 This Week's Schedule</h1>
      <p>Here's what's coming up this week.</p>
    </div>
    <div class="events">
      {{if .}}
        {{range .}}
        <div class="event">
          <h3>{{.Title}}</h3>
          <p>🕐 {{.StartDateTime.Format "Mon Jan 2, 3:04 PM"}}{{if not .EndDateTime.IsZero}} – {{.EndDateTime.Format "3:04 PM"}}{{end}}</p>
          {{if .Location}}<p>📍 {{.Location}}</p>{{end}}
        </div>
        {{end}}
      {{else}}
        <p class="empty">No events scheduled for this week. Check back soon!</p>
      {{end}}
    </div>
    <div class="footer">You're receiving this because you subscribed to the weekly schedule digest.</div>
  </div>
</body>
</html>`))

func buildEmailBody(events []models.EventItem) (string, error) {
	var buf bytes.Buffer
	if err := emailTmpl.Execute(&buf, events); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

package service

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP configuration.
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// EmailService sends emails via SMTP.
type EmailService struct {
	cfg EmailConfig
}

func NewEmailService(cfg EmailConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

// Send sends an email with the given subject and HTML body.
func (s *EmailService) Send(to, subject, body string) error {
	msg := strings.Join([]string{
		"From: " + s.cfg.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg))
}

// SendPasswordReset sends a password reset email.
func (s *EmailService) SendPasswordReset(to, resetURL string) error {
	body := fmt.Sprintf(`<p>Hello,</p>
<p>Someone has requested a link to change your password. You can do this through the link below.</p>
<p><a href="%s">Change my password</a></p>
<p>If you didn't request this, please ignore this email. Your password won't change until you access the link above and create a new one.</p>
<p>This link will expire in 6 hours.</p>`, resetURL)

	return s.Send(to, "Reset password instructions", body)
}

// SendConfirmation sends an email confirmation link.
func (s *EmailService) SendConfirmation(to, confirmURL string) error {
	body := fmt.Sprintf(`<p>Hello,</p>
<p>You can confirm your account email through the link below:</p>
<p><a href="%s">Confirm my account</a></p>
<p>If you didn't create an account, please ignore this email.</p>`, confirmURL)

	return s.Send(to, "Confirmation instructions", body)
}

// SendWelcome sends a welcome email after registration.
func (s *EmailService) SendWelcome(to, username string, needsApproval bool) error {
	var body string
	if needsApproval {
		body = fmt.Sprintf(`<p>Welcome %s,</p>
<p>Your account has been created and is pending approval by an administrator.</p>
<p>You will receive another email once your account has been activated.</p>`, username)
	} else {
		body = fmt.Sprintf(`<p>Welcome %s,</p>
<p>Your account has been created. Please check your email for confirmation instructions.</p>`, username)
	}

	return s.Send(to, "Welcome to Fat Free CRM", body)
}

// SendAssignmentNotification notifies a user they've been assigned to an entity.
func (s *EmailService) SendAssignmentNotification(to, assignerName, entityType, entityName, entityURL string) error {
	body := fmt.Sprintf(`<p>Hello,</p>
<p>%s has assigned you to the following %s:</p>
<p><strong><a href="%s">%s</a></strong></p>`, assignerName, entityType, entityURL, entityName)

	subject := fmt.Sprintf("Fat Free CRM: You have been assigned %s %s", entityName, entityType)
	return s.Send(to, subject, body)
}

// SendCommentNotification notifies subscribed users about a new comment.
// The subject includes [entity_type:id] so replies can be parsed as comments.
func (s *EmailService) SendCommentNotification(to, commenterName, entityType string, entityID int64, entityName, commentBody, entityURL string) error {
	body := fmt.Sprintf(`<p>%s commented on <a href="%s">%s</a>:</p>
<blockquote style="border-left: 3px solid #ccc; padding-left: 12px; color: #555;">%s</blockquote>
<p style="color: #888; font-size: 12px;">You can reply to this email to add a comment.</p>`,
		commenterName, entityURL, entityName, commentBody)

	subject := fmt.Sprintf("RE: [%s:%d] %s", entityType, entityID, entityName)
	return s.Send(to, subject, body)
}

// SendDropboxNotification notifies a user that an email was attached to entities.
func (s *EmailService) SendDropboxNotification(to, fromAddress, emailSubject string, entityLinks []string) error {
	links := ""
	for _, l := range entityLinks {
		links += fmt.Sprintf("<li>%s</li>", l)
	}

	body := fmt.Sprintf(`<p>Hello,</p>
<p>An email from <strong>%s</strong> with subject "<em>%s</em>" has been attached to:</p>
<ul>%s</ul>`, fromAddress, emailSubject, links)

	return s.Send(to, "Fat Free CRM: Email has been attached", body)
}

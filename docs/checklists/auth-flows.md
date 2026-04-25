### Auth Flows

**Password Reset**
- [x] POST /auth/forgot-password — generate reset token, send email
- [x] POST /auth/reset-password — validate token, set new password
- [x] Reset token with 6-hour expiry (matches Rails)
- [x] Tests

**User Registration**
- [x] POST /auth/register — create user (respects signup setting: allowed/needs_approval/not_allowed)
- [x] needs_approval: user created but suspended_at set, needs admin reactivation
- [x] Tests

**Email Confirmation**
- [x] POST /auth/confirm — validate confirmation token, confirm user
- [x] POST /auth/resend-confirmation — resend confirmation email
- [x] Tests

**SMTP Email Sending**
- [x] Email service using net/smtp
- [x] Password reset email template
- [x] Confirmation email template
- [x] Welcome/registration email template
- [x] Configurable SMTP settings (from env vars)

**React Frontend**
- [x] Forgot password page (enter email → check email message)
- [x] Reset password page (token from URL → new password form)
- [x] Registration page (if signup allowed)
- [x] Email confirmation page (token from URL → auto-confirm)
- [x] Links on login page (forgot password, sign up)

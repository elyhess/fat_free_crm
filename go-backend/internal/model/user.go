package model

import "time"

// User maps to the Rails `users` table.
type User struct {
	ID                   int64      `gorm:"primaryKey" json:"id"`
	Username             string     `gorm:"size:32" json:"username"`
	Email                string     `gorm:"size:254" json:"email"`
	FirstName            string     `gorm:"size:32" json:"first_name"`
	LastName             string     `gorm:"size:32" json:"last_name"`
	Title                string     `gorm:"size:64" json:"title,omitempty"`
	Company              string     `gorm:"size:64" json:"company,omitempty"`
	AltEmail             string     `gorm:"size:64" json:"alt_email,omitempty"`
	Phone                string     `gorm:"size:32" json:"phone,omitempty"`
	Mobile               string     `gorm:"size:32" json:"mobile,omitempty"`
	Admin                bool       `json:"admin"`
	EncryptedPassword    string     `gorm:"size:255;column:encrypted_password" json:"-"`
	PasswordSalt         string     `gorm:"size:255;column:password_salt" json:"-"`
	ResetPasswordToken   *string    `gorm:"column:reset_password_token" json:"-"`
	ResetPasswordSentAt  *time.Time `gorm:"column:reset_password_sent_at" json:"-"`
	ConfirmationToken    *string    `gorm:"size:255;column:confirmation_token" json:"-"`
	ConfirmedAt          *time.Time `json:"confirmed_at,omitempty"`
	ConfirmationSentAt   *time.Time `gorm:"column:confirmation_sent_at" json:"-"`
	UnconfirmedEmail     *string    `gorm:"size:254;column:unconfirmed_email" json:"-"`
	SuspendedAt          *time.Time `json:"suspended_at,omitempty"`
	SignInCount          int        `gorm:"column:sign_in_count" json:"sign_in_count"`
	CurrentSignInAt      *time.Time `gorm:"column:current_sign_in_at" json:"-"`
	LastSignInAt         *time.Time `gorm:"column:last_sign_in_at" json:"-"`
	CurrentSignInIP      string     `gorm:"size:255;column:current_sign_in_ip" json:"-"`
	LastSignInIP         string     `gorm:"size:255;column:last_sign_in_ip" json:"-"`
	DeletedAt            *time.Time `json:"-"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// IsConfirmed returns true if the user has confirmed their email.
func (u *User) IsConfirmed() bool {
	return u.ConfirmedAt != nil
}

// IsSuspended returns true if the user is suspended.
func (u *User) IsSuspended() bool {
	return u.SuspendedAt != nil
}

// IsActive returns true if the user can authenticate.
// Matches Rails: confirmed? && !suspended? (or suspended but has signed in before)
func (u *User) IsActive() bool {
	if !u.IsConfirmed() {
		return false
	}
	if u.IsSuspended() && u.SignInCount == 0 {
		return false // awaiting approval
	}
	if u.IsSuspended() {
		return false
	}
	return true
}

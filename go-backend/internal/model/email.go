package model

import "time"

// Email maps to the Rails `emails` table.
// Emails are polymorphically attached to entities via mediator_type/mediator_id.
type Email struct {
	ID            int64      `gorm:"primaryKey" json:"id"`
	ImapMessageID string     `gorm:"column:imap_message_id" json:"imap_message_id"`
	UserID        *int64     `json:"user_id,omitempty"`
	MediatorType  string     `gorm:"column:mediator_type" json:"mediator_type"`
	MediatorID    int64      `gorm:"column:mediator_id" json:"mediator_id"`
	SentFrom      string     `gorm:"column:sent_from" json:"sent_from"`
	SentTo        string     `gorm:"column:sent_to" json:"sent_to"`
	CC            string     `json:"cc,omitempty"`
	BCC           string     `json:"bcc,omitempty"`
	Subject       string     `json:"subject"`
	Body          string     `json:"body"`
	Header        string     `json:"-"`
	SentAt        *time.Time `gorm:"column:sent_at" json:"sent_at,omitempty"`
	ReceivedAt    *time.Time `gorm:"column:received_at" json:"received_at,omitempty"`
	DeletedAt     *time.Time `json:"-"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	State         string     `gorm:"default:Expanded" json:"state"`
}

func (Email) TableName() string { return "emails" }

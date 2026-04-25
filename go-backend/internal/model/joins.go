package model

import "time"

// AccountContact is the join table linking accounts to contacts.
type AccountContact struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AccountID int64     `json:"account_id"`
	ContactID int64     `json:"contact_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time
}

func (AccountContact) TableName() string { return "account_contacts" }

// AccountOpportunity is the join table linking accounts to opportunities.
type AccountOpportunity struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	AccountID     int64     `json:"account_id"`
	OpportunityID int64     `json:"opportunity_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     *time.Time
}

func (AccountOpportunity) TableName() string { return "account_opportunities" }

// ContactOpportunity is the join table linking contacts to opportunities.
type ContactOpportunity struct {
	ID            int64   `gorm:"primaryKey" json:"id"`
	ContactID     int64   `json:"contact_id"`
	OpportunityID int64   `json:"opportunity_id"`
	Role          *string `gorm:"size:32" json:"role,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     *time.Time
}

func (ContactOpportunity) TableName() string { return "contact_opportunities" }

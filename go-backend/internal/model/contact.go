package model

import (
	"time"
)

type Contact struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `json:"user_id"`
	LeadID         *int64     `json:"lead_id,omitempty"`
	AssignedTo     int64      `json:"assigned_to"`
	ReportsTo      *int64     `json:"reports_to,omitempty"`
	FirstName      string     `gorm:"size:64" json:"first_name"`
	LastName       string     `gorm:"size:64" json:"last_name"`
	Access         string     `gorm:"size:8" json:"access"`
	Title          *string    `gorm:"size:64" json:"title,omitempty"`
	Department     *string    `gorm:"size:64" json:"department,omitempty"`
	Email          *string    `gorm:"size:254" json:"email,omitempty"`
	AltEmail       *string    `gorm:"size:254" json:"alt_email,omitempty"`
	Phone          *string    `gorm:"size:32" json:"phone,omitempty"`
	Mobile         *string    `gorm:"size:32" json:"mobile,omitempty"`
	Fax            *string    `gorm:"size:32" json:"fax,omitempty"`
	Blog           *string    `gorm:"size:128" json:"blog,omitempty"`
	LinkedIn       *string    `gorm:"size:128" json:"linkedin,omitempty"`
	Facebook       *string    `gorm:"size:128" json:"facebook,omitempty"`
	Twitter        *string    `gorm:"size:128" json:"twitter,omitempty"`
	BornOn         *time.Time `json:"born_on,omitempty"`
	DoNotCall      bool       `json:"do_not_call"`
	BackgroundInfo *string    `json:"background_info,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

func (Contact) TableName() string { return "contacts" }

func (c Contact) GetID() int64         { return c.ID }
func (c Contact) GetAccess() string    { return c.Access }
func (c Contact) GetUserID() int64     { return c.UserID }
func (c Contact) GetAssignedTo() int64 { return c.AssignedTo }
func (c Contact) GetAssetType() string { return "Contact" }

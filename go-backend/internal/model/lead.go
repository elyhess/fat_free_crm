package model

import (
	"time"
)

type Lead struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `json:"user_id"`
	CampaignID     *int64     `json:"campaign_id,omitempty"`
	AssignedTo     int64      `json:"assigned_to"`
	FirstName      string     `gorm:"size:64" json:"first_name"`
	LastName       string     `gorm:"size:64" json:"last_name"`
	Access         string     `gorm:"size:8" json:"access"`
	Company        *string    `gorm:"size:64" json:"company,omitempty"`
	Title          *string    `gorm:"size:64" json:"title,omitempty"`
	Source         *string    `gorm:"size:32" json:"source,omitempty"`
	Status         *string    `gorm:"size:32" json:"status,omitempty"`
	ReferredBy     *string    `gorm:"size:64" json:"referred_by,omitempty"`
	Email          *string    `gorm:"size:254" json:"email,omitempty"`
	AltEmail       *string    `gorm:"size:254" json:"alt_email,omitempty"`
	Phone          *string    `gorm:"size:32" json:"phone,omitempty"`
	Mobile         *string    `gorm:"size:32" json:"mobile,omitempty"`
	Blog           *string    `gorm:"size:128" json:"blog,omitempty"`
	LinkedIn       *string    `gorm:"size:128" json:"linkedin,omitempty"`
	Facebook       *string    `gorm:"size:128" json:"facebook,omitempty"`
	Twitter        *string    `gorm:"size:128" json:"twitter,omitempty"`
	Rating         int        `json:"rating"`
	DoNotCall      bool       `json:"do_not_call"`
	BackgroundInfo *string    `json:"background_info,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

func (Lead) TableName() string { return "leads" }

func (l Lead) GetID() int64         { return l.ID }
func (l Lead) GetAccess() string    { return l.Access }
func (l Lead) GetUserID() int64     { return l.UserID }
func (l Lead) GetAssignedTo() int64 { return l.AssignedTo }
func (l Lead) GetAssetType() string { return "Lead" }

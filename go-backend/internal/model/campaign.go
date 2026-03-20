package model

import (
	"time"
)

type Campaign struct {
	ID                 int64      `gorm:"primaryKey" json:"id"`
	UserID             int64      `json:"user_id"`
	AssignedTo         int64      `json:"assigned_to"`
	Name               string     `gorm:"size:64" json:"name"`
	Access             string     `gorm:"size:8" json:"access"`
	Status             *string    `gorm:"size:64" json:"status,omitempty"`
	Budget             *float64   `gorm:"type:decimal(12,2)" json:"budget,omitempty"`
	TargetLeads        *int       `json:"target_leads,omitempty"`
	TargetConversion   *float64   `json:"target_conversion,omitempty"`
	TargetRevenue      *float64   `gorm:"type:decimal(12,2)" json:"target_revenue,omitempty"`
	LeadsCount         int        `json:"leads_count"`
	OpportunitiesCount int        `json:"opportunities_count"`
	Revenue            *float64   `gorm:"type:decimal(12,2)" json:"revenue,omitempty"`
	StartsOn           *time.Time `json:"starts_on,omitempty"`
	EndsOn             *time.Time `json:"ends_on,omitempty"`
	Objectives         *string    `json:"objectives,omitempty"`
	BackgroundInfo     *string    `json:"background_info,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

func (Campaign) TableName() string { return "campaigns" }

func (c Campaign) GetID() int64         { return c.ID }
func (c Campaign) GetAccess() string    { return c.Access }
func (c Campaign) GetUserID() int64     { return c.UserID }
func (c Campaign) GetAssignedTo() int64 { return c.AssignedTo }
func (c Campaign) GetAssetType() string { return "Campaign" }

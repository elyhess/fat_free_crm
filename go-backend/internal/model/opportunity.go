package model

import (
	"time"
)

type Opportunity struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `json:"user_id"`
	CampaignID     *int64     `json:"campaign_id,omitempty"`
	AssignedTo     int64      `json:"assigned_to"`
	Name           string     `gorm:"size:64" json:"name"`
	Access         string     `gorm:"size:8" json:"access"`
	Source         *string    `gorm:"size:32" json:"source,omitempty"`
	Stage          *string    `gorm:"size:32" json:"stage,omitempty"`
	Probability    *int       `json:"probability,omitempty"`
	Amount         *float64   `gorm:"type:decimal(12,2)" json:"amount,omitempty"`
	Discount       *float64   `gorm:"type:decimal(12,2)" json:"discount,omitempty"`
	ClosesOn       *time.Time `json:"closes_on,omitempty"`
	BackgroundInfo *string    `json:"background_info,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

func (Opportunity) TableName() string { return "opportunities" }

func (o Opportunity) GetID() int64         { return o.ID }
func (o Opportunity) GetAccess() string    { return o.Access }
func (o Opportunity) GetUserID() int64     { return o.UserID }
func (o Opportunity) GetAssignedTo() int64 { return o.AssignedTo }
func (o Opportunity) GetAssetType() string { return "Opportunity" }

// WeightedAmount calculates the weighted deal value.
func (o Opportunity) WeightedAmount() float64 {
	amt := 0.0
	if o.Amount != nil {
		amt = *o.Amount
	}
	disc := 0.0
	if o.Discount != nil {
		disc = *o.Discount
	}
	prob := 0
	if o.Probability != nil {
		prob = *o.Probability
	}
	return (amt - disc) * float64(prob) / 100.0
}

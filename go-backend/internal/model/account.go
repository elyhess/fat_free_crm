package model

import (
	"time"
)

type Account struct {
	ID                 int64      `gorm:"primaryKey" json:"id"`
	UserID             int64      `json:"user_id"`
	AssignedTo         int64      `json:"assigned_to"`
	Name               string     `gorm:"size:64" json:"name"`
	Access             string     `gorm:"size:8" json:"access"`
	Rating             int        `json:"rating"`
	Category           *string    `gorm:"size:32" json:"category,omitempty"`
	Email              *string    `gorm:"size:254" json:"email,omitempty"`
	Website            *string    `gorm:"size:128" json:"website,omitempty"`
	Phone              *string    `gorm:"size:32" json:"phone,omitempty"`
	TollFreePhone      *string    `gorm:"size:32" json:"toll_free_phone,omitempty"`
	Fax                *string    `gorm:"size:32" json:"fax,omitempty"`
	BackgroundInfo     *string    `json:"background_info,omitempty"`
	ContactsCount      int        `json:"contacts_count"`
	OpportunitiesCount int        `json:"opportunities_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

func (Account) TableName() string { return "accounts" }

func (a Account) GetID() int64         { return a.ID }
func (a Account) GetAccess() string    { return a.Access }
func (a Account) GetUserID() int64     { return a.UserID }
func (a Account) GetAssignedTo() int64 { return a.AssignedTo }
func (a Account) GetAssetType() string { return "Account" }

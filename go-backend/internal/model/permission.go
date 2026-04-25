package model

import "time"

// Permission maps to the Rails `permissions` table.
// Links users or groups to shared entity records.
type Permission struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    *int64    `json:"user_id,omitempty"`
	GroupID   *int64    `json:"group_id,omitempty"`
	AssetID   int64     `json:"asset_id"`
	AssetType string    `json:"asset_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Permission) TableName() string { return "permissions" }

// Group maps to the Rails `groups` table.
type Group struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Group) TableName() string { return "groups" }

// Access level constants matching Rails.
const (
	AccessPublic  = "Public"
	AccessPrivate = "Private"
	AccessShared  = "Shared"
)

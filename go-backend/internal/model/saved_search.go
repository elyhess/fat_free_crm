package model

import (
	"encoding/json"
	"time"
)

// SavedSearch stores a user's saved search/filter preset.
type SavedSearch struct {
	ID        int64           `gorm:"primaryKey" json:"id"`
	UserID    int64           `json:"user_id"`
	Name      string          `json:"name"`
	Entity    string          `json:"entity"`
	Filters   json.RawMessage `gorm:"type:jsonb" json:"filters"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (SavedSearch) TableName() string { return "saved_searches" }

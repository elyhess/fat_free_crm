package model

import "time"

// Tag maps to the acts-as-taggable-on `tags` table.
type Tag struct {
	ID            int64  `gorm:"primaryKey" json:"id"`
	Name          string `json:"name"`
	TaggingsCount int    `json:"taggings_count"`
}

func (Tag) TableName() string { return "tags" }

// Tagging maps to the acts-as-taggable-on `taggings` join table.
type Tagging struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	TagID        int64     `json:"tag_id"`
	TaggableID   int64     `json:"taggable_id"`
	TaggableType string    `gorm:"size:50" json:"taggable_type"`
	TaggerID     *int64    `json:"tagger_id,omitempty"`
	TaggerType   *string   `json:"tagger_type,omitempty"`
	Context      string    `gorm:"size:50" json:"context"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Tagging) TableName() string { return "taggings" }

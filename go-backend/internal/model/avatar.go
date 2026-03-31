package model

import "time"

// Avatar maps to the Rails `avatars` table.
type Avatar struct {
	ID               int64     `gorm:"primaryKey" json:"id"`
	UserID           int64     `json:"user_id"`
	EntityType       string    `json:"entity_type"`
	EntityID         int64     `json:"entity_id"`
	ImageFileName    *string   `gorm:"column:image_file_name" json:"image_file_name"`
	ImageContentType *string   `gorm:"column:image_content_type" json:"image_content_type"`
	ImageFileSize    *int64    `gorm:"column:image_file_size" json:"image_file_size"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Avatar) TableName() string { return "avatars" }

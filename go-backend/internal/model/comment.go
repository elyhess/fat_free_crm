package model

import "time"

// Comment maps to the Rails `comments` table (polymorphic).
type Comment struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	UserID          int64     `json:"user_id"`
	CommentableID   int64     `json:"commentable_id"`
	CommentableType string    `json:"commentable_type"`
	Private         bool      `json:"private"`
	Title           string    `json:"title"`
	Comment         string    `json:"comment"`
	State           string    `gorm:"size:16" json:"state"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Comment) TableName() string { return "comments" }

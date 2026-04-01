package model

import "time"

// ResearchTool maps to the Rails `research_tools` table.
type ResearchTool struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	URLTemplate string    `gorm:"column:url_template" json:"url_template"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ResearchTool) TableName() string { return "research_tools" }

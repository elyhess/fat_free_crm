package model

import (
	"time"
)

type Task struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `json:"user_id"`
	AssignedTo     int64      `json:"assigned_to"`
	CompletedBy    *int64     `json:"completed_by,omitempty"`
	Name           string     `gorm:"size:255" json:"name"`
	AssetID        *int64     `json:"asset_id,omitempty"`
	AssetType      *string    `gorm:"size:255" json:"asset_type,omitempty"`
	Priority       *string    `gorm:"size:32" json:"priority,omitempty"`
	Category       *string    `gorm:"size:32" json:"category,omitempty"`
	Bucket         *string    `gorm:"size:32" json:"bucket,omitempty"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	BackgroundInfo *string    `json:"background_info,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

func (Task) TableName() string { return "tasks" }

// IsCompleted returns true if the task has been completed.
func (t Task) IsCompleted() bool { return t.CompletedAt != nil }

// Task buckets matching Rails constants.
const (
	BucketDueASAP      = "due_asap"
	BucketDueToday     = "due_today"
	BucketDueTomorrow  = "due_tomorrow"
	BucketDueThisWeek  = "due_this_week"
	BucketDueNextWeek  = "due_next_week"
	BucketDueLater     = "due_later"
	BucketOverdue      = "overdue"
	BucketSpecificTime = "specific_time"
)

package model

import "time"

// Version maps to the PaperTrail `versions` table (audit log).
type Version struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	ItemType      string    `json:"item_type"`
	ItemID        int64     `json:"item_id"`
	Event         string    `gorm:"size:512" json:"event"`
	Whodunnit     *string   `json:"whodunnit,omitempty"`
	Object        *string   `json:"object,omitempty"`
	ObjectChanges *string   `json:"object_changes,omitempty"`
	RelatedID     *int64    `json:"related_id,omitempty"`
	RelatedType   *string   `json:"related_type,omitempty"`
	TransactionID *int64    `json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func (Version) TableName() string { return "versions" }

// Tracked entity types matching Rails ASSETS constant.
var TrackedAssets = []string{"Account", "Campaign", "Contact", "Lead", "Opportunity"}

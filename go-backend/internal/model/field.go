package model

import "time"

// FieldGroup maps to the Rails `field_groups` table.
// Each group belongs to an entity type (klass_name) and contains ordered fields.
type FieldGroup struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64" json:"name"`
	Label     string    `gorm:"size:128" json:"label"`
	Position  int       `json:"position"`
	Hint      string    `json:"hint,omitempty"`
	KlassName string    `gorm:"size:32;column:klass_name" json:"klass_name"`
	TagID     *int64    `json:"tag_id,omitempty"`
	Fields    []Field   `gorm:"foreignKey:FieldGroupID" json:"fields"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (FieldGroup) TableName() string { return "field_groups" }

// Field maps to the Rails `fields` table.
// Uses STI via the Type column (CoreField, CustomField, CustomFieldPair, etc.).
// The `as` column determines the UI rendering and DB column type.
type Field struct {
	ID           int64   `gorm:"primaryKey" json:"id"`
	Type         string  `gorm:"column:type;size:255" json:"type"`
	FieldGroupID int64   `gorm:"column:field_group_id" json:"field_group_id"`
	Position     int     `json:"position"`
	Name         string  `gorm:"size:64" json:"name"`
	Label        string  `gorm:"size:128" json:"label"`
	Hint         string  `json:"hint,omitempty"`
	Placeholder  string  `json:"placeholder,omitempty"`
	As           string  `gorm:"column:as;size:32" json:"as"`
	Collection   string  `gorm:"type:text" json:"collection,omitempty"`
	Disabled     bool    `json:"disabled"`
	Required     bool    `json:"required"`
	Maxlength    *int    `json:"maxlength,omitempty"`
	Minlength    int     `gorm:"default:0" json:"minlength"`
	PairID       *int64  `json:"pair_id,omitempty"`
	Settings     string  `gorm:"type:text" json:"settings,omitempty"`
	Pattern      string  `json:"pattern,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Field) TableName() string { return "fields" }

// ValidEntityTypes are the entity class names that support custom fields.
var ValidEntityTypes = map[string]string{
	"Account":     "accounts",
	"Contact":     "contacts",
	"Lead":        "leads",
	"Opportunity": "opportunities",
	"Campaign":    "campaigns",
	"Task":        "tasks",
}

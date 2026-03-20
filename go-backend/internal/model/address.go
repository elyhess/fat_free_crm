package model

import "time"

// Address maps to the Rails `addresses` table (polymorphic).
type Address struct {
	ID              int64      `gorm:"primaryKey" json:"id"`
	Street1         *string    `json:"street1,omitempty"`
	Street2         *string    `json:"street2,omitempty"`
	City            *string    `gorm:"size:64" json:"city,omitempty"`
	State           *string    `gorm:"size:64" json:"state,omitempty"`
	Zipcode         *string    `gorm:"size:16" json:"zipcode,omitempty"`
	Country         *string    `gorm:"size:64" json:"country,omitempty"`
	FullAddress     *string    `json:"full_address,omitempty"`
	AddressType     *string    `gorm:"size:16" json:"address_type,omitempty"`
	AddressableID   int64      `json:"addressable_id"`
	AddressableType string     `json:"addressable_type"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

func (Address) TableName() string { return "addresses" }

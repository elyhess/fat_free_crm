package service

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// VersionRecorder creates PaperTrail-compatible version records for entity mutations.
type VersionRecorder struct {
	db *gorm.DB
}

func NewVersionRecorder(db *gorm.DB) *VersionRecorder {
	return &VersionRecorder{db: db}
}

// RecordCreate writes a version record for a newly created entity.
func (v *VersionRecorder) RecordCreate(itemType string, itemID int64, userID int64, object interface{}) {
	objJSON := marshalObject(object)
	whodunnit := fmt.Sprintf("%d", userID)
	ver := model.Version{
		ItemType:  itemType,
		ItemID:    itemID,
		Event:     "create",
		Whodunnit: &whodunnit,
		Object:    objJSON,
		CreatedAt: time.Now().UTC(),
	}
	v.db.Create(&ver)
}

// RecordUpdate writes a version record for an entity update.
// oldObject is the state before the update; changes is the map of field->new value that was applied.
func (v *VersionRecorder) RecordUpdate(itemType string, itemID int64, userID int64, oldObject interface{}, changes map[string]interface{}) {
	objJSON := marshalObject(oldObject)
	changesJSON := marshalObject(changes)
	whodunnit := fmt.Sprintf("%d", userID)
	ver := model.Version{
		ItemType:      itemType,
		ItemID:        itemID,
		Event:         "update",
		Whodunnit:     &whodunnit,
		Object:        objJSON,
		ObjectChanges: changesJSON,
		CreatedAt:     time.Now().UTC(),
	}
	v.db.Create(&ver)
}

// RecordDestroy writes a version record for a deleted entity.
func (v *VersionRecorder) RecordDestroy(itemType string, itemID int64, userID int64, object interface{}) {
	objJSON := marshalObject(object)
	whodunnit := fmt.Sprintf("%d", userID)
	ver := model.Version{
		ItemType:  itemType,
		ItemID:    itemID,
		Event:     "destroy",
		Whodunnit: &whodunnit,
		Object:    objJSON,
		CreatedAt: time.Now().UTC(),
	}
	v.db.Create(&ver)
}

func marshalObject(obj interface{}) *string {
	if obj == nil {
		return nil
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

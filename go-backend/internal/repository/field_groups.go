package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

type FieldGroupRepository struct {
	db *gorm.DB
}

func NewFieldGroupRepository(db *gorm.DB) *FieldGroupRepository {
	return &FieldGroupRepository{db: db}
}

// FindByEntity returns all field groups for an entity type with their fields, ordered by position.
func (r *FieldGroupRepository) FindByEntity(klassName string) ([]model.FieldGroup, error) {
	if _, ok := model.ValidEntityTypes[klassName]; !ok {
		return nil, fmt.Errorf("invalid entity type: %s", klassName)
	}

	var groups []model.FieldGroup
	err := r.db.
		Where("klass_name = ?", klassName).
		Order("position ASC").
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Find(&groups).Error

	return groups, err
}

// FindFieldsByEntity returns all fields for an entity type, ordered by group position then field position.
func (r *FieldGroupRepository) FindFieldsByEntity(klassName string) ([]model.Field, error) {
	if _, ok := model.ValidEntityTypes[klassName]; !ok {
		return nil, fmt.Errorf("invalid entity type: %s", klassName)
	}

	var fields []model.Field
	err := r.db.
		Joins("JOIN field_groups ON field_groups.id = fields.field_group_id").
		Where("field_groups.klass_name = ?", klassName).
		Order("field_groups.position ASC, fields.position ASC").
		Find(&fields).Error

	return fields, err
}

// FindCustomFieldsByEntity returns only CustomField-type fields (the ones with cf_* columns).
func (r *FieldGroupRepository) FindCustomFieldsByEntity(klassName string) ([]model.Field, error) {
	if _, ok := model.ValidEntityTypes[klassName]; !ok {
		return nil, fmt.Errorf("invalid entity type: %s", klassName)
	}

	var fields []model.Field
	err := r.db.
		Joins("JOIN field_groups ON field_groups.id = fields.field_group_id").
		Where("field_groups.klass_name = ? AND fields.type != ?", klassName, "CoreField").
		Order("field_groups.position ASC, fields.position ASC").
		Find(&fields).Error

	return fields, err
}

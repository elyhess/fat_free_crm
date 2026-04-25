package repository

import (
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// SupportingRepository handles reads for comments, addresses, tags, and versions.
type SupportingRepository struct {
	db *gorm.DB
}

func NewSupportingRepository(db *gorm.DB) *SupportingRepository {
	return &SupportingRepository{db: db}
}

// ListComments returns comments for a given entity (polymorphic).
func (r *SupportingRepository) ListComments(entityType string, entityID int64) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.
		Where("commentable_type = ? AND commentable_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}

// ListAddresses returns addresses for a given entity (polymorphic).
func (r *SupportingRepository) ListAddresses(entityType string, entityID int64) ([]model.Address, error) {
	var addresses []model.Address
	err := r.db.
		Where("addressable_type = ? AND addressable_id = ? AND deleted_at IS NULL", entityType, entityID).
		Order("address_type").
		Find(&addresses).Error
	return addresses, err
}

// ListTagsForEntity returns tags attached to a given entity.
func (r *SupportingRepository) ListTagsForEntity(entityType string, entityID int64) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.
		Joins("JOIN taggings ON taggings.tag_id = tags.id").
		Where("taggings.taggable_type = ? AND taggings.taggable_id = ?", entityType, entityID).
		Find(&tags).Error
	return tags, err
}

// ListAllTags returns all tags ordered by name.
func (r *SupportingRepository) ListAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Order("name").Find(&tags).Error
	return tags, err
}

// ListVersions returns recent audit log entries for a given entity.
func (r *SupportingRepository) ListVersions(entityType string, entityID int64, limit int) ([]model.Version, error) {
	var versions []model.Version
	err := r.db.
		Where("item_type = ? AND item_id = ?", entityType, entityID).
		Order("created_at DESC").
		Limit(limit).
		Find(&versions).Error
	return versions, err
}

// ListRecentActivity returns recent versions across all tracked assets.
func (r *SupportingRepository) ListRecentActivity(limit int) ([]model.Version, error) {
	var versions []model.Version
	err := r.db.
		Where("item_type IN ?", model.TrackedAssets).
		Order("created_at DESC").
		Limit(limit).
		Find(&versions).Error
	return versions, err
}

// ListUsers returns all non-deleted users (admin endpoint).
func (r *SupportingRepository) ListUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.
		Where("deleted_at IS NULL").
		Order("username").
		Find(&users).Error
	return users, err
}

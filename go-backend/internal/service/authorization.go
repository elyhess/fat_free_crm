package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// AuthorizationService handles access control decisions.
type AuthorizationService struct {
	db *gorm.DB
}

func NewAuthorizationService(db *gorm.DB) *AuthorizationService {
	return &AuthorizationService{db: db}
}

// CanAccess checks if a user can access a specific entity record.
// The entity must have access, user_id, and assigned_to fields.
func (s *AuthorizationService) CanAccess(userID int64, isAdmin bool, record EntityRecord) (bool, error) {
	if isAdmin {
		return true, nil
	}

	switch record.GetAccess() {
	case model.AccessPublic:
		return true, nil
	case model.AccessPrivate:
		return record.GetUserID() == userID || record.GetAssignedTo() == userID, nil
	case model.AccessShared:
		if record.GetUserID() == userID || record.GetAssignedTo() == userID {
			return true, nil
		}
		return s.hasPermission(userID, record.GetID(), record.GetAssetType())
	default:
		return false, nil
	}
}

// EntityRecord is the interface that entity models must implement for access control.
type EntityRecord interface {
	GetID() int64
	GetAccess() string
	GetUserID() int64
	GetAssignedTo() int64
	GetAssetType() string
}

// hasPermission checks the permissions table for direct user or group-based access.
func (s *AuthorizationService) hasPermission(userID, assetID int64, assetType string) (bool, error) {
	// Get user's group IDs
	var groupIDs []int64
	s.db.Table("groups_users").
		Where("user_id = ?", userID).
		Pluck("group_id", &groupIDs)

	query := s.db.Model(&model.Permission{}).
		Where("asset_id = ? AND asset_type = ?", assetID, assetType)

	if len(groupIDs) > 0 {
		query = query.Where("user_id = ? OR group_id IN ?", userID, groupIDs)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ScopeAccessible returns a GORM scope that filters records to only those
// accessible by the given user. Use this for list endpoints.
//
// Usage: db.Scopes(authSvc.ScopeAccessible(userID, isAdmin, "Account")).Find(&accounts)
func (s *AuthorizationService) ScopeAccessible(userID int64, isAdmin bool, assetType string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isAdmin {
			return db
		}

		tableName, ok := model.ValidEntityTypes[assetType]
		if !ok {
			return db.Where("1 = 0") // invalid type, return nothing
		}

		// Get user's group IDs for shared access check
		var groupIDs []int64
		s.db.Table("groups_users").
			Where("user_id = ?", userID).
			Pluck("group_id", &groupIDs)

		// Build the permission subquery for Shared records
		permSubquery := fmt.Sprintf(
			"EXISTS (SELECT 1 FROM permissions WHERE permissions.asset_id = %s.id AND permissions.asset_type = ? AND (permissions.user_id = ?",
			tableName,
		)
		args := []interface{}{assetType, userID}

		if len(groupIDs) > 0 {
			permSubquery += " OR permissions.group_id IN ?"
			args = append(args, groupIDs)
		}
		permSubquery += "))"

		// Combine: Public OR owner OR assignee OR (Shared + has permission)
		allArgs := []interface{}{userID, userID}
		allArgs = append(allArgs, args...)

		return db.Where(
			fmt.Sprintf(
				"%s.access = 'Public' OR %s.user_id = ? OR %s.assigned_to = ? OR (%s.access = 'Shared' AND %s)",
				tableName, tableName, tableName, tableName, permSubquery,
			),
			allArgs...,
		)
	}
}

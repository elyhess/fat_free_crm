package repository

import (
	"strings"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByLogin looks up a user by username or email (case-insensitive).
// Matches Rails: where("lower(username) = :value OR lower(email) = :value", value: login.downcase)
func (r *UserRepository) FindByLogin(login string) (*model.User, error) {
	login = strings.ToLower(strings.TrimSpace(login))
	var user model.User
	err := r.db.
		Where("lower(username) = ? OR lower(email) = ?", login, login).
		Where("deleted_at IS NULL").
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID looks up a user by ID.
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.
		Where("deleted_at IS NULL").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateSignInTracking updates sign-in tracking fields after successful login.
func (r *UserRepository) UpdateSignInTracking(user *model.User, ip string) error {
	return r.db.Model(user).Updates(map[string]interface{}{
		"sign_in_count":       gorm.Expr("sign_in_count + 1"),
		"last_sign_in_at":     user.CurrentSignInAt,
		"last_sign_in_ip":     user.CurrentSignInIP,
		"current_sign_in_at":  gorm.Expr("NOW()"),
		"current_sign_in_ip":  ip,
	}).Error
}

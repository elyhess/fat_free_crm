package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// allowedSortColumns defines valid sort columns per entity table to prevent SQL injection.
var allowedSortColumns = map[string]map[string]bool{
	"tasks": {
		"id": true, "name": true, "created_at": true, "updated_at": true,
		"due_at": true, "completed_at": true, "bucket": true, "priority": true,
	},
	"accounts": {
		"id": true, "name": true, "created_at": true, "updated_at": true,
		"rating": true, "category": true,
	},
	"contacts": {
		"id": true, "first_name": true, "last_name": true, "created_at": true,
		"updated_at": true,
	},
	"leads": {
		"id": true, "first_name": true, "last_name": true, "created_at": true,
		"updated_at": true, "status": true, "rating": true,
	},
	"opportunities": {
		"id": true, "name": true, "created_at": true, "updated_at": true,
		"stage": true, "amount": true, "closes_on": true, "probability": true,
	},
	"campaigns": {
		"id": true, "name": true, "created_at": true, "updated_at": true,
		"status": true, "starts_on": true, "ends_on": true,
	},
}

// EntityRepository provides generic CRUD operations for CRM entities.
type EntityRepository[T any] struct {
	db        *gorm.DB
	tableName string
}

// NewEntityRepository creates a repository for the given entity type.
func NewEntityRepository[T any](db *gorm.DB, tableName string) *EntityRepository[T] {
	return &EntityRepository[T]{db: db, tableName: tableName}
}

// List returns a paginated list of entities, applying the given scopes.
func (r *EntityRepository[T]) List(params model.PaginationParams, scopes ...func(*gorm.DB) *gorm.DB) (*model.PaginatedResult[T], error) {
	var total int64
	query := r.db.Table(r.tableName).Where("deleted_at IS NULL")
	for _, s := range scopes {
		query = query.Scopes(s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Validate and apply sort
	sortCol := "id"
	if cols, ok := allowedSortColumns[r.tableName]; ok {
		if cols[params.Sort] {
			sortCol = params.Sort
		}
	}
	order := "DESC"
	if params.Order == "asc" {
		order = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage
	var items []T
	err := query.
		Order(fmt.Sprintf("%s.%s %s", r.tableName, sortCol, order)).
		Limit(params.PerPage).
		Offset(offset).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / params.PerPage
	if int(total)%params.PerPage > 0 {
		totalPages++
	}

	return &model.PaginatedResult[T]{
		Data:       items,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// FindByID returns a single entity by ID, applying the given scopes.
func (r *EntityRepository[T]) FindByID(id int64, scopes ...func(*gorm.DB) *gorm.DB) (*T, error) {
	var item T
	query := r.db.Table(r.tableName).Where("deleted_at IS NULL").Where("id = ?", id)
	for _, s := range scopes {
		query = query.Scopes(s)
	}
	if err := query.First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

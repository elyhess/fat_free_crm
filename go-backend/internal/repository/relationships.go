package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// RelationshipRepository provides queries for entity-to-entity relationships.
type RelationshipRepository struct {
	db *gorm.DB
}

func NewRelationshipRepository(db *gorm.DB) *RelationshipRepository {
	return &RelationshipRepository{db: db}
}

// ListContactsForAccount returns contacts linked to an account via the account_contacts join table.
func (r *RelationshipRepository) ListContactsForAccount(accountID int64, params model.PaginationParams, scopes ...func(*gorm.DB) *gorm.DB) (*model.PaginatedResult[model.Contact], error) {
	query := r.joinQuery("contacts", "account_contacts", "contact_id", "account_id", accountID, scopes...)
	return paginateQuery[model.Contact](query, "contacts", params)
}

// ListOpportunitiesForAccount returns opportunities linked to an account via the account_opportunities join table.
func (r *RelationshipRepository) ListOpportunitiesForAccount(accountID int64, params model.PaginationParams, scopes ...func(*gorm.DB) *gorm.DB) (*model.PaginatedResult[model.Opportunity], error) {
	query := r.joinQuery("opportunities", "account_opportunities", "opportunity_id", "account_id", accountID, scopes...)
	return paginateQuery[model.Opportunity](query, "opportunities", params)
}

// ListOpportunitiesForContact returns opportunities linked to a contact via the contact_opportunities join table.
func (r *RelationshipRepository) ListOpportunitiesForContact(contactID int64, params model.PaginationParams, scopes ...func(*gorm.DB) *gorm.DB) (*model.PaginatedResult[model.Opportunity], error) {
	query := r.joinQuery("opportunities", "contact_opportunities", "opportunity_id", "contact_id", contactID, scopes...)
	return paginateQuery[model.Opportunity](query, "opportunities", params)
}

// ListLeadsForCampaign returns leads belonging to a campaign via the campaign_id foreign key.
func (r *RelationshipRepository) ListLeadsForCampaign(campaignID int64, params model.PaginationParams, scopes ...func(*gorm.DB) *gorm.DB) (*model.PaginatedResult[model.Lead], error) {
	query := r.fkQuery("leads", "campaign_id", campaignID, scopes...)
	return paginateQuery[model.Lead](query, "leads", params)
}

// ListOpportunitiesForCampaign returns opportunities belonging to a campaign via the campaign_id foreign key.
func (r *RelationshipRepository) ListOpportunitiesForCampaign(campaignID int64, params model.PaginationParams, scopes ...func(*gorm.DB) *gorm.DB) (*model.PaginatedResult[model.Opportunity], error) {
	query := r.fkQuery("opportunities", "campaign_id", campaignID, scopes...)
	return paginateQuery[model.Opportunity](query, "opportunities", params)
}

// joinQuery builds a query for related entities through a join table.
func (r *RelationshipRepository) joinQuery(
	entityTable, joinTable, joinEntityCol, joinOwnerCol string,
	ownerID int64,
	scopes ...func(*gorm.DB) *gorm.DB,
) *gorm.DB {
	joinClause := fmt.Sprintf(
		"JOIN %s ON %s.%s = %s.id AND %s.%s = ? AND %s.deleted_at IS NULL",
		joinTable, joinTable, joinEntityCol, entityTable, joinTable, joinOwnerCol, joinTable,
	)

	query := r.db.Table(entityTable).
		Joins(joinClause, ownerID).
		Where(fmt.Sprintf("%s.deleted_at IS NULL", entityTable))

	for _, s := range scopes {
		query = query.Scopes(s)
	}
	return query
}

// fkQuery builds a query for related entities through a direct foreign key.
func (r *RelationshipRepository) fkQuery(
	entityTable, fkColumn string,
	ownerID int64,
	scopes ...func(*gorm.DB) *gorm.DB,
) *gorm.DB {
	query := r.db.Table(entityTable).
		Where(fmt.Sprintf("%s.%s = ?", entityTable, fkColumn), ownerID).
		Where(fmt.Sprintf("%s.deleted_at IS NULL", entityTable))

	for _, s := range scopes {
		query = query.Scopes(s)
	}
	return query
}

// paginateQuery applies count, sort, offset, and limit to a query and returns a PaginatedResult.
func paginateQuery[T any](query *gorm.DB, tableName string, params model.PaginationParams) (*model.PaginatedResult[T], error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "id"
	if cols, ok := allowedSortColumns[tableName]; ok {
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
		Order(fmt.Sprintf("%s.%s %s", tableName, sortCol, order)).
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

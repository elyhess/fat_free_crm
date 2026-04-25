package handler

import (
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// SearchHandler provides search endpoints across entities.
type SearchHandler struct {
	db       *gorm.DB
	authzSvc *service.AuthorizationService
}

func NewSearchHandler(db *gorm.DB, authzSvc *service.AuthorizationService) *SearchHandler {
	return &SearchHandler{db: db, authzSvc: authzSvc}
}

// SearchResult holds results grouped by entity type.
type SearchResult struct {
	Query         string              `json:"query"`
	Accounts      []model.Account     `json:"accounts"`
	Contacts      []model.Contact     `json:"contacts"`
	Leads         []model.Lead        `json:"leads"`
	Opportunities []model.Opportunity `json:"opportunities"`
	Campaigns     []model.Campaign    `json:"campaigns"`
	Tasks         []model.Task        `json:"tasks"`
	TotalCount    int                 `json:"total_count"`
}

// toTSQuery converts a user search string into a PostgreSQL tsquery.
// Multi-word queries become prefix-matched terms joined with &.
// Example: "john smith" -> "john:* & smith:*"
func toTSQuery(q string) string {
	words := strings.Fields(q)
	if len(words) == 0 {
		return ""
	}
	parts := make([]string, len(words))
	for i, w := range words {
		// Escape single quotes and use prefix matching
		safe := strings.ReplaceAll(w, "'", "''")
		parts[i] = safe + ":*"
	}
	return strings.Join(parts, " & ")
}

// Search performs a cross-entity search using PostgreSQL full-text search
// with ILIKE fallback for entities without tsvector columns.
// GET /api/v1/search?q=term&entity=accounts (entity filter is optional)
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	entityFilter := strings.ToLower(r.URL.Query().Get("entity"))
	limit := 25
	tsq := toTSQuery(q)
	pattern := "%" + q + "%" // ILIKE fallback

	result := SearchResult{Query: q}

	if entityFilter == "" || entityFilter == "accounts" {
		var accounts []model.Account
		scope := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Account"))
		if tsq != "" {
			scope = scope.Where("accounts.tsv @@ to_tsquery('english', ?)", tsq).
				Order(fmt.Sprintf("ts_rank(accounts.tsv, to_tsquery('english', %s)) DESC",
					quoteLiteral(tsq)))
		} else {
			scope = scope.Where("accounts.name ILIKE ? OR accounts.email ILIKE ?", pattern, pattern)
		}
		scope.Limit(limit).Find(&accounts)
		result.Accounts = accounts
	}

	if entityFilter == "" || entityFilter == "contacts" {
		var contacts []model.Contact
		scope := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Contact"))
		if tsq != "" {
			scope = scope.Where("contacts.tsv @@ to_tsquery('english', ?)", tsq).
				Order(fmt.Sprintf("ts_rank(contacts.tsv, to_tsquery('english', %s)) DESC",
					quoteLiteral(tsq)))
		} else {
			scope = scope.Where("contacts.first_name ILIKE ? OR contacts.last_name ILIKE ? OR contacts.email ILIKE ?", pattern, pattern, pattern)
		}
		scope.Limit(limit).Find(&contacts)
		result.Contacts = contacts
	}

	if entityFilter == "" || entityFilter == "leads" {
		var leads []model.Lead
		scope := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Lead"))
		if tsq != "" {
			scope = scope.Where("leads.tsv @@ to_tsquery('english', ?)", tsq).
				Order(fmt.Sprintf("ts_rank(leads.tsv, to_tsquery('english', %s)) DESC",
					quoteLiteral(tsq)))
		} else {
			scope = scope.Where("leads.first_name ILIKE ? OR leads.last_name ILIKE ? OR leads.company ILIKE ? OR leads.email ILIKE ?", pattern, pattern, pattern, pattern)
		}
		scope.Limit(limit).Find(&leads)
		result.Leads = leads
	}

	if entityFilter == "" || entityFilter == "opportunities" {
		var opps []model.Opportunity
		scope := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Opportunity"))
		if tsq != "" {
			scope = scope.Where("opportunities.tsv @@ to_tsquery('english', ?)", tsq).
				Order(fmt.Sprintf("ts_rank(opportunities.tsv, to_tsquery('english', %s)) DESC",
					quoteLiteral(tsq)))
		} else {
			scope = scope.Where("opportunities.name ILIKE ?", pattern)
		}
		scope.Limit(limit).Find(&opps)
		result.Opportunities = opps
	}

	if entityFilter == "" || entityFilter == "campaigns" {
		var campaigns []model.Campaign
		scope := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Campaign"))
		if tsq != "" {
			scope = scope.Where("campaigns.tsv @@ to_tsquery('english', ?)", tsq).
				Order(fmt.Sprintf("ts_rank(campaigns.tsv, to_tsquery('english', %s)) DESC",
					quoteLiteral(tsq)))
		} else {
			scope = scope.Where("campaigns.name ILIKE ?", pattern)
		}
		scope.Limit(limit).Find(&campaigns)
		result.Campaigns = campaigns
	}

	// Tasks don't have tsvector — use ILIKE
	if entityFilter == "" || entityFilter == "tasks" {
		var tasks []model.Task
		h.db.Where("(tasks.user_id = ? OR tasks.assigned_to = ?) AND tasks.name ILIKE ?",
			claims.UserID, claims.UserID, pattern).
			Limit(limit).Find(&tasks)
		result.Tasks = tasks
	}

	result.TotalCount = len(result.Accounts) + len(result.Contacts) + len(result.Leads) +
		len(result.Opportunities) + len(result.Campaigns) + len(result.Tasks)

	writeJSON(w, http.StatusOK, result)
}

// quoteLiteral returns a single-quoted PostgreSQL string literal for use in ORDER BY.
func quoteLiteral(s string) string {
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
}

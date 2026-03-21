package handler

import (
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

// Search performs a cross-entity search.
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
	pattern := "%" + q + "%"
	limit := 25

	result := SearchResult{Query: q}

	if entityFilter == "" || entityFilter == "accounts" {
		var accounts []model.Account
		h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Account")).
			Where("accounts.name LIKE ? OR accounts.email LIKE ?", pattern, pattern).
			Limit(limit).Find(&accounts)
		result.Accounts = accounts
	}

	if entityFilter == "" || entityFilter == "contacts" {
		var contacts []model.Contact
		h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Contact")).
			Where("contacts.first_name LIKE ? OR contacts.last_name LIKE ? OR contacts.email LIKE ?", pattern, pattern, pattern).
			Limit(limit).Find(&contacts)
		result.Contacts = contacts
	}

	if entityFilter == "" || entityFilter == "leads" {
		var leads []model.Lead
		h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Lead")).
			Where("leads.first_name LIKE ? OR leads.last_name LIKE ? OR leads.company LIKE ? OR leads.email LIKE ?", pattern, pattern, pattern, pattern).
			Limit(limit).Find(&leads)
		result.Leads = leads
	}

	if entityFilter == "" || entityFilter == "opportunities" {
		var opps []model.Opportunity
		h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Opportunity")).
			Where("opportunities.name LIKE ?", pattern).
			Limit(limit).Find(&opps)
		result.Opportunities = opps
	}

	if entityFilter == "" || entityFilter == "campaigns" {
		var campaigns []model.Campaign
		h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Campaign")).
			Where("campaigns.name LIKE ?", pattern).
			Limit(limit).Find(&campaigns)
		result.Campaigns = campaigns
	}

	if entityFilter == "" || entityFilter == "tasks" {
		var tasks []model.Task
		h.db.Where("(tasks.user_id = ? OR tasks.assigned_to = ?) AND tasks.name LIKE ?",
			claims.UserID, claims.UserID, pattern).
			Limit(limit).Find(&tasks)
		result.Tasks = tasks
	}

	result.TotalCount = len(result.Accounts) + len(result.Contacts) + len(result.Leads) +
		len(result.Opportunities) + len(result.Campaigns) + len(result.Tasks)

	writeJSON(w, http.StatusOK, result)
}

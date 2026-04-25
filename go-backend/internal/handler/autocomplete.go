package handler

import (
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// AutocompleteHandler provides typeahead search endpoints for entity pickers.
type AutocompleteHandler struct {
	db       *gorm.DB
	authzSvc *service.AuthorizationService
}

func NewAutocompleteHandler(db *gorm.DB, authzSvc *service.AuthorizationService) *AutocompleteHandler {
	return &AutocompleteHandler{db: db, authzSvc: authzSvc}
}

type autocompleteItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Accounts returns matching accounts.
// GET /accounts/autocomplete?q=...
func (h *AutocompleteHandler) Accounts(w http.ResponseWriter, r *http.Request) {
	h.search(w, r, "accounts", "Account", "name")
}

// Contacts returns matching contacts (first_name || ' ' || last_name).
// GET /contacts/autocomplete?q=...
func (h *AutocompleteHandler) Contacts(w http.ResponseWriter, r *http.Request) {
	h.searchConcat(w, r, "contacts", "Contact", "first_name", "last_name")
}

// Leads returns matching leads (first_name || ' ' || last_name).
// GET /leads/autocomplete?q=...
func (h *AutocompleteHandler) Leads(w http.ResponseWriter, r *http.Request) {
	h.searchConcat(w, r, "leads", "Lead", "first_name", "last_name")
}

// Campaigns returns matching campaigns.
// GET /campaigns/autocomplete?q=...
func (h *AutocompleteHandler) Campaigns(w http.ResponseWriter, r *http.Request) {
	h.search(w, r, "campaigns", "Campaign", "name")
}

// Opportunities returns matching opportunities.
// GET /opportunities/autocomplete?q=...
func (h *AutocompleteHandler) Opportunities(w http.ResponseWriter, r *http.Request) {
	h.search(w, r, "opportunities", "Opportunity", "name")
}

// search handles autocomplete for entities with a single name column.
func (h *AutocompleteHandler) search(w http.ResponseWriter, r *http.Request, table, assetType, nameCol string) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []autocompleteItem{})
		return
	}

	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, assetType)

	var results []autocompleteItem
	query := h.db.Table(table).
		Select(fmt.Sprintf("id, %s AS name", nameCol)).
		Where(fmt.Sprintf("deleted_at IS NULL AND %s ILIKE ?", nameCol), "%"+q+"%").
		Scopes(scope).
		Order(nameCol).
		Limit(10)

	if err := query.Scan(&results).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if results == nil {
		results = []autocompleteItem{}
	}
	writeJSON(w, http.StatusOK, results)
}

// searchConcat handles autocomplete for entities with first_name + last_name.
func (h *AutocompleteHandler) searchConcat(w http.ResponseWriter, r *http.Request, table, assetType, firstCol, lastCol string) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []autocompleteItem{})
		return
	}

	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, assetType)
	nameExpr := fmt.Sprintf("%s || ' ' || %s", firstCol, lastCol)

	var results []autocompleteItem
	query := h.db.Table(table).
		Select(fmt.Sprintf("id, %s AS name", nameExpr)).
		Where(fmt.Sprintf("deleted_at IS NULL AND (%s ILIKE ? OR %s ILIKE ?)", firstCol, lastCol), "%"+q+"%", "%"+q+"%").
		Scopes(scope).
		Order(firstCol).
		Limit(10)

	if err := query.Scan(&results).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if results == nil {
		results = []autocompleteItem{}
	}
	writeJSON(w, http.StatusOK, results)
}

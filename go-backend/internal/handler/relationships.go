package handler

import (
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// RelationshipHandler serves endpoints for entity-to-entity relationships.
type RelationshipHandler struct {
	repo     *repository.RelationshipRepository
	authzSvc *service.AuthorizationService
}

func NewRelationshipHandler(db *gorm.DB, authzSvc *service.AuthorizationService) *RelationshipHandler {
	return &RelationshipHandler{
		repo:     repository.NewRelationshipRepository(db),
		authzSvc: authzSvc,
	}
}

// AccountContacts returns contacts linked to an account.
// GET /api/v1/accounts/{id}/contacts
func (h *RelationshipHandler) AccountContacts(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := h.parseRequest(w, r)
	if !ok {
		return
	}
	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Contact")
	result, err := h.repo.ListContactsForAccount(id, parsePagination(r), scope)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// AccountOpportunities returns opportunities linked to an account.
// GET /api/v1/accounts/{id}/opportunities
func (h *RelationshipHandler) AccountOpportunities(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := h.parseRequest(w, r)
	if !ok {
		return
	}
	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Opportunity")
	result, err := h.repo.ListOpportunitiesForAccount(id, parsePagination(r), scope)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// CampaignLeads returns leads belonging to a campaign.
// GET /api/v1/campaigns/{id}/leads
func (h *RelationshipHandler) CampaignLeads(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := h.parseRequest(w, r)
	if !ok {
		return
	}
	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Lead")
	result, err := h.repo.ListLeadsForCampaign(id, parsePagination(r), scope)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// CampaignOpportunities returns opportunities belonging to a campaign.
// GET /api/v1/campaigns/{id}/opportunities
func (h *RelationshipHandler) CampaignOpportunities(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := h.parseRequest(w, r)
	if !ok {
		return
	}
	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Opportunity")
	result, err := h.repo.ListOpportunitiesForCampaign(id, parsePagination(r), scope)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ContactOpportunities returns opportunities linked to a contact.
// GET /api/v1/contacts/{id}/opportunities
func (h *RelationshipHandler) ContactOpportunities(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := h.parseRequest(w, r)
	if !ok {
		return
	}
	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Opportunity")
	result, err := h.repo.ListOpportunitiesForContact(id, parsePagination(r), scope)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// parseRequest extracts JWT claims and the {id} path parameter.
func (h *RelationshipHandler) parseRequest(w http.ResponseWriter, r *http.Request) (*auth.Claims, int64, bool) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, 0, false
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return nil, 0, false
	}
	return claims, id, true
}


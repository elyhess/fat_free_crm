package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// entityHandler provides list and detail endpoints for a CRM entity.
type entityHandler[T any] struct {
	repo     *repository.EntityRepository[T]
	authzSvc *service.AuthorizationService
	asset    string // e.g. "Account", "Task"
}

func parsePagination(r *http.Request) model.PaginationParams {
	p := model.DefaultPagination()
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Page = n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			p.PerPage = n
		}
	}
	if v := r.URL.Query().Get("sort"); v != "" {
		p.Sort = v
	}
	if v := r.URL.Query().Get("order"); v != "" {
		p.Order = v
	}
	return p
}

func (h *entityHandler[T]) list(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	params := parsePagination(r)
	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, h.asset)

	result, err := h.repo.List(params, scope)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *entityHandler[T]) detail(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, h.asset)
	item, err := h.repo.FindByID(id, scope)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// RegisterEntityRoutes wires up list and detail routes for all CRM entities.
func RegisterEntityRoutes(mux interface {
	Get(pattern string, handler http.HandlerFunc)
}, db *gorm.DB, authzSvc *service.AuthorizationService) {
	// Tasks (no access field — uses different access model)
	tasks := &entityHandler[model.Task]{
		repo:     repository.NewEntityRepository[model.Task](db, "tasks"),
		authzSvc: authzSvc,
		asset:    "Task",
	}
	mux.Get("/tasks", tasks.list)
	mux.Get("/tasks/{id}", tasks.detail)

	// Accounts
	accounts := &entityHandler[model.Account]{
		repo:     repository.NewEntityRepository[model.Account](db, "accounts"),
		authzSvc: authzSvc,
		asset:    "Account",
	}
	mux.Get("/accounts", accounts.list)
	mux.Get("/accounts/{id}", accounts.detail)

	// Contacts
	contacts := &entityHandler[model.Contact]{
		repo:     repository.NewEntityRepository[model.Contact](db, "contacts"),
		authzSvc: authzSvc,
		asset:    "Contact",
	}
	mux.Get("/contacts", contacts.list)
	mux.Get("/contacts/{id}", contacts.detail)

	// Leads
	leads := &entityHandler[model.Lead]{
		repo:     repository.NewEntityRepository[model.Lead](db, "leads"),
		authzSvc: authzSvc,
		asset:    "Lead",
	}
	mux.Get("/leads", leads.list)
	mux.Get("/leads/{id}", leads.detail)

	// Opportunities
	opps := &entityHandler[model.Opportunity]{
		repo:     repository.NewEntityRepository[model.Opportunity](db, "opportunities"),
		authzSvc: authzSvc,
		asset:    "Opportunity",
	}
	mux.Get("/opportunities", opps.list)
	mux.Get("/opportunities/{id}", opps.detail)

	// Campaigns
	campaigns := &entityHandler[model.Campaign]{
		repo:     repository.NewEntityRepository[model.Campaign](db, "campaigns"),
		authzSvc: authzSvc,
		asset:    "Campaign",
	}
	mux.Get("/campaigns", campaigns.list)
	mux.Get("/campaigns/{id}", campaigns.detail)
}

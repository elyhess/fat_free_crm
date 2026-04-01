package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// filterableColumns defines which columns can be filtered per entity table.
var filterableColumns = map[string]map[string]string{
	"accounts": {
		"name": "string", "email": "string", "rating": "int",
		"category": "string", "access": "string",
	},
	"contacts": {
		"first_name": "string", "last_name": "string", "email": "string",
		"department": "string", "source": "string", "access": "string",
	},
	"leads": {
		"first_name": "string", "last_name": "string", "email": "string",
		"company": "string", "status": "string", "rating": "int",
		"source": "string", "access": "string",
	},
	"opportunities": {
		"name": "string", "stage": "string", "amount": "decimal",
		"probability": "int", "source": "string", "access": "string",
	},
	"campaigns": {
		"name": "string", "status": "string", "access": "string",
	},
	"tasks": {
		"name": "string", "bucket": "string", "priority": "string",
		"category": "string",
	},
}

// parseFilters extracts filter[field_op]=value query params and returns a GORM scope.
// Supported operators: eq, cont (contains/ILIKE), gt, lt, blank, present.
// Default operator (no suffix) is "cont" for strings, "eq" for others.
func parseFilters(r *http.Request, tableName string) func(*gorm.DB) *gorm.DB {
	allowed := filterableColumns[tableName]
	if allowed == nil {
		return func(db *gorm.DB) *gorm.DB { return db }
	}

	type filter struct {
		column string
		op     string
		value  string
	}
	var filters []filter

	for key, values := range r.URL.Query() {
		if !strings.HasPrefix(key, "filter[") || !strings.HasSuffix(key, "]") {
			continue
		}
		inner := key[7 : len(key)-1] // strip "filter[" and "]"
		val := values[0]

		// Parse field and operator: "name_cont" -> field="name", op="cont"
		field, op := inner, ""
		for _, suffix := range []string{"_eq", "_cont", "_gt", "_lt", "_blank", "_present"} {
			if strings.HasSuffix(inner, suffix) {
				field = inner[:len(inner)-len(suffix)]
				op = suffix[1:] // strip leading underscore
				break
			}
		}

		colType, ok := allowed[field]
		if !ok {
			continue
		}

		// Default operator
		if op == "" {
			if colType == "string" {
				op = "cont"
			} else {
				op = "eq"
			}
		}

		filters = append(filters, filter{column: field, op: op, value: val})
	}

	return func(db *gorm.DB) *gorm.DB {
		for _, f := range filters {
			col := fmt.Sprintf("%s.%s", tableName, f.column)
			switch f.op {
			case "eq":
				db = db.Where(col+" = ?", f.value)
			case "cont":
				db = db.Where(col+" ILIKE ?", "%"+f.value+"%")
			case "gt":
				db = db.Where(col+" > ?", f.value)
			case "lt":
				db = db.Where(col+" < ?", f.value)
			case "blank":
				db = db.Where(col+" IS NULL OR "+col+" = ''")
			case "present":
				db = db.Where(col+" IS NOT NULL AND "+col+" != ''")
			}
		}
		return db
	}
}

func (h *entityHandler[T]) list(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	params := parsePagination(r)
	scope := h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, h.asset)
	filterScope := parseFilters(r, h.repo.TableName())

	result, err := h.repo.List(params, scope, filterScope)
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

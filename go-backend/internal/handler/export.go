package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// ExportHandler provides CSV export endpoints.
type ExportHandler struct {
	db       *gorm.DB
	authzSvc *service.AuthorizationService
}

func NewExportHandler(db *gorm.DB, authzSvc *service.AuthorizationService) *ExportHandler {
	return &ExportHandler{db: db, authzSvc: authzSvc}
}

// ExportAccounts exports accounts as CSV.
// GET /api/v1/accounts/export
func (h *ExportHandler) ExportAccounts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var accounts []model.Account
	query := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Account"))
	query.Order("id ASC").Find(&accounts)

	headers := []string{"ID", "Name", "Email", "Phone", "Website", "Fax", "Category", "Rating", "Access", "Created At"}
	rows := make([][]string, 0, len(accounts))
	for _, a := range accounts {
		rows = append(rows, []string{
			i64(a.ID), a.Name, s(a.Email), s(a.Phone), s(a.Website), s(a.Fax),
			s(a.Category), intStr(a.Rating), a.Access, fmtTime(a.CreatedAt),
		})
	}
	writeCSV(w, "accounts", headers, rows)
}

// ExportContacts exports contacts as CSV.
// GET /api/v1/contacts/export
func (h *ExportHandler) ExportContacts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var contacts []model.Contact
	query := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Contact"))
	query.Order("id ASC").Find(&contacts)

	headers := []string{"ID", "First Name", "Last Name", "Title", "Department", "Email", "Phone", "Mobile", "Access", "Created At"}
	rows := make([][]string, 0, len(contacts))
	for _, c := range contacts {
		rows = append(rows, []string{
			i64(c.ID), c.FirstName, c.LastName, sptr(c.Title), sptr(c.Department),
			sptr(c.Email), sptr(c.Phone), sptr(c.Mobile), c.Access, fmtTime(c.CreatedAt),
		})
	}
	writeCSV(w, "contacts", headers, rows)
}

// ExportLeads exports leads as CSV.
// GET /api/v1/leads/export
func (h *ExportHandler) ExportLeads(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var leads []model.Lead
	query := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Lead"))
	query.Order("id ASC").Find(&leads)

	headers := []string{"ID", "First Name", "Last Name", "Company", "Title", "Status", "Email", "Phone", "Source", "Rating", "Access", "Created At"}
	rows := make([][]string, 0, len(leads))
	for _, l := range leads {
		rows = append(rows, []string{
			i64(l.ID), l.FirstName, l.LastName, sptr(l.Company), sptr(l.Title),
			sptr(l.Status), sptr(l.Email), sptr(l.Phone), sptr(l.Source),
			intStr(l.Rating), l.Access, fmtTime(l.CreatedAt),
		})
	}
	writeCSV(w, "leads", headers, rows)
}

// ExportOpportunities exports opportunities as CSV.
// GET /api/v1/opportunities/export
func (h *ExportHandler) ExportOpportunities(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var opps []model.Opportunity
	query := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Opportunity"))
	query.Order("id ASC").Find(&opps)

	headers := []string{"ID", "Name", "Stage", "Amount", "Probability", "Discount", "Closes On", "Source", "Access", "Created At"}
	rows := make([][]string, 0, len(opps))
	for _, o := range opps {
		closesOn := ""
		if o.ClosesOn != nil {
			closesOn = o.ClosesOn.Format("2006-01-02")
		}
		rows = append(rows, []string{
			i64(o.ID), o.Name, sptr(o.Stage), decStr(o.Amount), intPtrStr(o.Probability),
			decStr(o.Discount), closesOn, sptr(o.Source), o.Access, fmtTime(o.CreatedAt),
		})
	}
	writeCSV(w, "opportunities", headers, rows)
}

// ExportCampaigns exports campaigns as CSV.
// GET /api/v1/campaigns/export
func (h *ExportHandler) ExportCampaigns(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var campaigns []model.Campaign
	query := h.db.Scopes(h.authzSvc.ScopeAccessible(claims.UserID, claims.Admin, "Campaign"))
	query.Order("id ASC").Find(&campaigns)

	headers := []string{"ID", "Name", "Status", "Budget", "Target Leads", "Leads Count", "Starts On", "Ends On", "Access", "Created At"}
	rows := make([][]string, 0, len(campaigns))
	for _, c := range campaigns {
		startsOn, endsOn := "", ""
		if c.StartsOn != nil {
			startsOn = c.StartsOn.Format("2006-01-02")
		}
		if c.EndsOn != nil {
			endsOn = c.EndsOn.Format("2006-01-02")
		}
		rows = append(rows, []string{
			i64(c.ID), c.Name, sptr(c.Status), decStr(c.Budget),
			intPtrStr(c.TargetLeads), intStr(c.LeadsCount),
			startsOn, endsOn, c.Access, fmtTime(c.CreatedAt),
		})
	}
	writeCSV(w, "campaigns", headers, rows)
}

// ExportTasks exports tasks as CSV.
// GET /api/v1/tasks/export
func (h *ExportHandler) ExportTasks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var tasks []model.Task
	h.db.Where("user_id = ? OR assigned_to = ?", claims.UserID, claims.UserID).
		Order("id ASC").Find(&tasks)

	headers := []string{"ID", "Name", "Priority", "Category", "Bucket", "Due At", "Completed At", "Created At"}
	rows := make([][]string, 0, len(tasks))
	for _, t := range tasks {
		dueAt, completedAt := "", ""
		if t.DueAt != nil {
			dueAt = t.DueAt.Format(time.RFC3339)
		}
		if t.CompletedAt != nil {
			completedAt = t.CompletedAt.Format(time.RFC3339)
		}
		rows = append(rows, []string{
			i64(t.ID), t.Name, sptr(t.Priority), sptr(t.Category), sptr(t.Bucket),
			dueAt, completedAt, fmtTime(t.CreatedAt),
		})
	}
	writeCSV(w, "tasks", headers, rows)
}

// --- CSV helpers ---

func writeCSV(w http.ResponseWriter, name string, headers []string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, name))

	writer := csv.NewWriter(w)
	_ = writer.Write(headers)
	for _, row := range rows {
		_ = writer.Write(row)
	}
	writer.Flush()
}

func i64(v int64) string              { return strconv.FormatInt(v, 10) }
func intStr(v int) string             { return strconv.Itoa(v) }
func fmtTime(t time.Time) string      { return t.Format(time.RFC3339) }

func sptr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func s(v *string) string {
	return sptr(v)
}

func intPtrStr(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}

func decStr(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

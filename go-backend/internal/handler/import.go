package handler

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// ImportHandler provides CSV import endpoints.
type ImportHandler struct {
	db *gorm.DB
}

func NewImportHandler(db *gorm.DB) *ImportHandler {
	return &ImportHandler{db: db}
}

// ImportAccounts imports accounts from CSV.
// POST /api/v1/accounts/import
func (h *ImportHandler) ImportAccounts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	records, headers, err := parseCSV(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSV: "+err.Error())
		return
	}

	colMap := mapColumns(headers)
	var created int
	var errors []string

	for i, row := range records {
		name := getField(row, colMap, "name")
		if name == "" {
			errors = append(errors, "row "+strconv.Itoa(i+2)+": name is required")
			continue
		}

		acct := model.Account{
			UserID:     claims.UserID,
			AssignedTo: claims.UserID,
			Name:       name,
			Access:     "Public",
		}
		if v := getField(row, colMap, "email"); v != "" {
			acct.Email = &v
		}
		if v := getField(row, colMap, "phone"); v != "" {
			acct.Phone = &v
		}
		if v := getField(row, colMap, "website"); v != "" {
			acct.Website = &v
		}
		if v := getField(row, colMap, "category"); v != "" {
			acct.Category = &v
		}
		if v := getField(row, colMap, "access"); v != "" {
			acct.Access = v
		}
		if v := getField(row, colMap, "rating"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				acct.Rating = n
			}
		}

		if err := h.db.Create(&acct).Error; err != nil {
			errors = append(errors, "row "+strconv.Itoa(i+2)+": "+err.Error())
			continue
		}
		created++
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"imported": created,
		"total":    len(records),
		"errors":   errors,
	})
}

// ImportContacts imports contacts from CSV.
// POST /api/v1/contacts/import
func (h *ImportHandler) ImportContacts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	records, headers, err := parseCSV(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSV: "+err.Error())
		return
	}

	colMap := mapColumns(headers)
	var created int
	var errors []string

	for i, row := range records {
		firstName := getField(row, colMap, "first_name", "first name", "firstname")
		lastName := getField(row, colMap, "last_name", "last name", "lastname")
		if firstName == "" && lastName == "" {
			errors = append(errors, "row "+strconv.Itoa(i+2)+": first_name or last_name is required")
			continue
		}

		contact := model.Contact{
			UserID:     claims.UserID,
			AssignedTo: claims.UserID,
			FirstName:  firstName,
			LastName:   lastName,
			Access:     "Public",
		}
		if v := getField(row, colMap, "email"); v != "" {
			contact.Email = &v
		}
		if v := getField(row, colMap, "phone"); v != "" {
			contact.Phone = &v
		}
		if v := getField(row, colMap, "mobile"); v != "" {
			contact.Mobile = &v
		}
		if v := getField(row, colMap, "title"); v != "" {
			contact.Title = &v
		}
		if v := getField(row, colMap, "department"); v != "" {
			contact.Department = &v
		}
		if v := getField(row, colMap, "access"); v != "" {
			contact.Access = v
		}

		if err := h.db.Create(&contact).Error; err != nil {
			errors = append(errors, "row "+strconv.Itoa(i+2)+": "+err.Error())
			continue
		}
		created++
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"imported": created,
		"total":    len(records),
		"errors":   errors,
	})
}

// ImportLeads imports leads from CSV.
// POST /api/v1/leads/import
func (h *ImportHandler) ImportLeads(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	records, headers, err := parseCSV(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSV: "+err.Error())
		return
	}

	colMap := mapColumns(headers)
	var created int
	var errors []string

	for i, row := range records {
		firstName := getField(row, colMap, "first_name", "first name", "firstname")
		lastName := getField(row, colMap, "last_name", "last name", "lastname")
		if firstName == "" && lastName == "" {
			errors = append(errors, "row "+strconv.Itoa(i+2)+": first_name or last_name is required")
			continue
		}

		lead := model.Lead{
			UserID:     claims.UserID,
			AssignedTo: claims.UserID,
			FirstName:  firstName,
			LastName:   lastName,
			Access:     "Public",
		}
		if v := getField(row, colMap, "company"); v != "" {
			lead.Company = &v
		}
		if v := getField(row, colMap, "email"); v != "" {
			lead.Email = &v
		}
		if v := getField(row, colMap, "phone"); v != "" {
			lead.Phone = &v
		}
		if v := getField(row, colMap, "title"); v != "" {
			lead.Title = &v
		}
		if v := getField(row, colMap, "source"); v != "" {
			lead.Source = &v
		}
		if v := getField(row, colMap, "status"); v != "" {
			lead.Status = &v
		}
		if v := getField(row, colMap, "access"); v != "" {
			lead.Access = v
		}

		if err := h.db.Create(&lead).Error; err != nil {
			errors = append(errors, "row "+strconv.Itoa(i+2)+": "+err.Error())
			continue
		}
		created++
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"imported": created,
		"total":    len(records),
		"errors":   errors,
	})
}

// --- CSV parsing helpers ---

func parseCSV(r *http.Request) ([][]string, []string, error) {
	// Support both multipart file upload and raw CSV body
	var reader io.Reader
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		file, _, err := r.FormFile("file")
		if err != nil {
			return nil, nil, err
		}
		defer func() { _ = file.Close() }()
		reader = file
	} else {
		reader = r.Body
	}

	csvReader := csv.NewReader(reader)
	allRecords, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(allRecords) < 1 {
		return nil, nil, nil
	}

	headers := allRecords[0]
	records := allRecords[1:]
	return records, headers, nil
}

// mapColumns creates a case-insensitive header-name → column-index map.
func mapColumns(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[normalize(h)] = i
	}
	return m
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// getField retrieves a value from a row by trying multiple header aliases.
func getField(row []string, colMap map[string]int, aliases ...string) string {
	for _, alias := range aliases {
		if idx, ok := colMap[normalize(alias)]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
	}
	return ""
}

// ImportResult is returned as JSON after import.
type ImportResult struct {
	Imported int      `json:"imported"`
	Total    int      `json:"total"`
	Errors   []string `json:"errors"`
}

// ExportFieldMapping returns the expected CSV headers for an entity type.
// GET /api/v1/{entity}/import/template
func (h *ImportHandler) ImportTemplate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entity := r.PathValue("entity")
	var headers []string

	switch entity {
	case "accounts":
		headers = []string{"Name", "Email", "Phone", "Website", "Category", "Rating", "Access"}
	case "contacts":
		headers = []string{"First Name", "Last Name", "Title", "Department", "Email", "Phone", "Mobile", "Access"}
	case "leads":
		headers = []string{"First Name", "Last Name", "Company", "Title", "Email", "Phone", "Source", "Status", "Access"}
	default:
		writeError(w, http.StatusBadRequest, "unsupported entity for import")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="`+entity+`_template.csv"`)
	writer := csv.NewWriter(w)
	_ = writer.Write(headers)
	writer.Flush()
}

// VCardExportContacts exports contacts as vCard format.
// GET /api/v1/contacts/export/vcard
func (h *ImportHandler) VCardExportContacts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Reuse the db from import handler
	var contacts []model.Contact
	h.db.Where("deleted_at IS NULL").Order("id ASC").Find(&contacts)

	w.Header().Set("Content-Type", "text/vcard")
	w.Header().Set("Content-Disposition", `attachment; filename="contacts.vcf"`)

	for _, c := range contacts {
		writeVCard(w, c)
	}
}

func writeVCard(w http.ResponseWriter, c model.Contact) {
	lines := []string{
		"BEGIN:VCARD",
		"VERSION:3.0",
		"N:" + c.LastName + ";" + c.FirstName + ";;;",
		"FN:" + c.FirstName + " " + c.LastName,
	}
	if c.Email != nil && *c.Email != "" {
		lines = append(lines, "EMAIL;TYPE=INTERNET:"+*c.Email)
	}
	if c.Phone != nil && *c.Phone != "" {
		lines = append(lines, "TEL;TYPE=WORK:"+*c.Phone)
	}
	if c.Mobile != nil && *c.Mobile != "" {
		lines = append(lines, "TEL;TYPE=CELL:"+*c.Mobile)
	}
	if c.Title != nil && *c.Title != "" {
		lines = append(lines, "TITLE:"+*c.Title)
	}
	if c.Department != nil && *c.Department != "" {
		lines = append(lines, "ORG:;"+*c.Department)
	}
	lines = append(lines, "END:VCARD")

	for _, line := range lines {
		_, _ = w.Write([]byte(line + "\r\n"))
	}
}

// ensure ImportResult implements json marshaling
var _ json.Marshaler = (*ImportResult)(nil)

func (r *ImportResult) MarshalJSON() ([]byte, error) {
	type Alias ImportResult
	return json.Marshal((*Alias)(r))
}

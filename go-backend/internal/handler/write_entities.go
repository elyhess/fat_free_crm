package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// --- Generic entity write helpers ---

// entityWriteAccess checks if the user can modify a record (admin, owner, or assignee).
func entityWriteAccess(claims *auth.Claims, record service.EntityRecord) bool {
	if claims.Admin {
		return true
	}
	return record.GetUserID() == claims.UserID || record.GetAssignedTo() == claims.UserID
}

// --- Accounts ---

type accountRequest struct {
	Name           string  `json:"name"`
	AssignedTo     int64   `json:"assigned_to"`
	Access         string  `json:"access"`
	Rating         int     `json:"rating"`
	Category       *string `json:"category,omitempty"`
	Email          *string `json:"email,omitempty"`
	Website        *string `json:"website,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	TollFreePhone  *string `json:"toll_free_phone,omitempty"`
	Fax            *string `json:"fax,omitempty"`
	BackgroundInfo *string `json:"background_info,omitempty"`
}

func (h *WriteHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req accountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	if req.Access == "" {
		req.Access = model.AccessPublic
	}

	acct := model.Account{
		UserID:         claims.UserID,
		AssignedTo:     req.AssignedTo,
		Name:           req.Name,
		Access:         req.Access,
		Rating:         req.Rating,
		Category:       req.Category,
		Email:          req.Email,
		Website:        req.Website,
		Phone:          req.Phone,
		TollFreePhone:  req.TollFreePhone,
		Fax:            req.Fax,
		BackgroundInfo: req.BackgroundInfo,
	}

	if err := h.db.Create(&acct).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}
	writeJSON(w, http.StatusCreated, acct)
}

func (h *WriteHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var acct model.Account
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&acct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, acct) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req accountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Access != "" {
		updates["access"] = req.Access
	}
	if req.AssignedTo != 0 {
		updates["assigned_to"] = req.AssignedTo
	}
	if req.Rating != 0 {
		updates["rating"] = req.Rating
	}
	if req.Category != nil {
		updates["category"] = req.Category
	}
	if req.Email != nil {
		updates["email"] = req.Email
	}
	if req.Website != nil {
		updates["website"] = req.Website
	}
	if req.Phone != nil {
		updates["phone"] = req.Phone
	}
	if req.BackgroundInfo != nil {
		updates["background_info"] = req.BackgroundInfo
	}

	if len(updates) > 0 {
		h.db.Model(&acct).Updates(updates)
	}

	h.db.First(&acct, id)
	writeJSON(w, http.StatusOK, acct)
}

func (h *WriteHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var acct model.Account
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&acct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, acct) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	h.db.Delete(&acct)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Campaigns ---

type campaignRequest struct {
	Name             string   `json:"name"`
	AssignedTo       int64    `json:"assigned_to"`
	Access           string   `json:"access"`
	Status           *string  `json:"status,omitempty"`
	Budget           *float64 `json:"budget,omitempty"`
	TargetLeads      *int     `json:"target_leads,omitempty"`
	TargetConversion *float64 `json:"target_conversion,omitempty"`
	TargetRevenue    *float64 `json:"target_revenue,omitempty"`
	StartsOn         *string  `json:"starts_on,omitempty"`
	EndsOn           *string  `json:"ends_on,omitempty"`
	Objectives       *string  `json:"objectives,omitempty"`
	BackgroundInfo   *string  `json:"background_info,omitempty"`
}

func (h *WriteHandler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req campaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	if req.Access == "" {
		req.Access = model.AccessPublic
	}

	campaign := model.Campaign{
		UserID:           claims.UserID,
		AssignedTo:       req.AssignedTo,
		Name:             req.Name,
		Access:           req.Access,
		Status:           req.Status,
		Budget:           req.Budget,
		TargetLeads:      req.TargetLeads,
		TargetConversion: req.TargetConversion,
		TargetRevenue:    req.TargetRevenue,
		Objectives:       req.Objectives,
		BackgroundInfo:   req.BackgroundInfo,
	}
	if req.StartsOn != nil {
		t, err := time.Parse("2006-01-02", *req.StartsOn)
		if err == nil {
			campaign.StartsOn = &t
		}
	}
	if req.EndsOn != nil {
		t, err := time.Parse("2006-01-02", *req.EndsOn)
		if err == nil {
			campaign.EndsOn = &t
		}
	}

	if err := h.db.Create(&campaign).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create campaign")
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (h *WriteHandler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var campaign model.Campaign
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, campaign) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req campaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Access != "" {
		updates["access"] = req.Access
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Budget != nil {
		updates["budget"] = *req.Budget
	}
	if req.Objectives != nil {
		updates["objectives"] = *req.Objectives
	}
	if req.BackgroundInfo != nil {
		updates["background_info"] = *req.BackgroundInfo
	}

	if len(updates) > 0 {
		h.db.Model(&campaign).Updates(updates)
	}

	h.db.First(&campaign, id)
	writeJSON(w, http.StatusOK, campaign)
}

func (h *WriteHandler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var campaign model.Campaign
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, campaign) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	h.db.Delete(&campaign)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Leads ---

type leadRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	AssignedTo     int64   `json:"assigned_to"`
	Access         string  `json:"access"`
	Company        *string `json:"company,omitempty"`
	Title          *string `json:"title,omitempty"`
	Source         *string `json:"source,omitempty"`
	Status         *string `json:"status,omitempty"`
	ReferredBy     *string `json:"referred_by,omitempty"`
	Email          *string `json:"email,omitempty"`
	AltEmail       *string `json:"alt_email,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	Mobile         *string `json:"mobile,omitempty"`
	Rating         int     `json:"rating"`
	DoNotCall      bool    `json:"do_not_call"`
	CampaignID     *int64  `json:"campaign_id,omitempty"`
	BackgroundInfo *string `json:"background_info,omitempty"`
}

func (h *WriteHandler) CreateLead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req leadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.LastName == "" {
		writeError(w, http.StatusUnprocessableEntity, "last_name is required")
		return
	}
	if req.Access == "" {
		req.Access = model.AccessPublic
	}

	lead := model.Lead{
		UserID:         claims.UserID,
		AssignedTo:     req.AssignedTo,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Access:         req.Access,
		Company:        req.Company,
		Title:          req.Title,
		Source:         req.Source,
		Status:         req.Status,
		ReferredBy:     req.ReferredBy,
		Email:          req.Email,
		AltEmail:       req.AltEmail,
		Phone:          req.Phone,
		Mobile:         req.Mobile,
		Rating:         req.Rating,
		DoNotCall:      req.DoNotCall,
		CampaignID:     req.CampaignID,
		BackgroundInfo: req.BackgroundInfo,
	}

	if err := h.db.Create(&lead).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create lead")
		return
	}

	// Increment campaign leads_count
	if lead.CampaignID != nil {
		h.db.Exec("UPDATE campaigns SET leads_count = leads_count + 1 WHERE id = ?", *lead.CampaignID)
	}

	writeJSON(w, http.StatusCreated, lead)
}

func (h *WriteHandler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var lead model.Lead
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&lead).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, lead) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req leadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Access != "" {
		updates["access"] = req.Access
	}
	if req.Company != nil {
		updates["company"] = *req.Company
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.BackgroundInfo != nil {
		updates["background_info"] = *req.BackgroundInfo
	}

	if len(updates) > 0 {
		h.db.Model(&lead).Updates(updates)
	}

	h.db.First(&lead, id)
	writeJSON(w, http.StatusOK, lead)
}

func (h *WriteHandler) DeleteLead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var lead model.Lead
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&lead).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, lead) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Decrement campaign leads_count
	if lead.CampaignID != nil {
		h.db.Exec("UPDATE campaigns SET leads_count = CASE WHEN leads_count > 0 THEN leads_count - 1 ELSE 0 END WHERE id = ?", *lead.CampaignID)
	}

	h.db.Delete(&lead)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *WriteHandler) RejectLead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var lead model.Lead
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&lead).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, lead) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	rejected := "rejected"
	h.db.Model(&lead).Update("status", rejected)
	lead.Status = &rejected
	writeJSON(w, http.StatusOK, lead)
}

// --- Contacts ---

type contactRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	AssignedTo     int64   `json:"assigned_to"`
	Access         string  `json:"access"`
	Title          *string `json:"title,omitempty"`
	Department     *string `json:"department,omitempty"`
	Email          *string `json:"email,omitempty"`
	AltEmail       *string `json:"alt_email,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	Mobile         *string `json:"mobile,omitempty"`
	Fax            *string `json:"fax,omitempty"`
	DoNotCall      bool    `json:"do_not_call"`
	BackgroundInfo *string `json:"background_info,omitempty"`
}

func (h *WriteHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.LastName == "" {
		writeError(w, http.StatusUnprocessableEntity, "last_name is required")
		return
	}
	if req.Access == "" {
		req.Access = model.AccessPublic
	}

	contact := model.Contact{
		UserID:         claims.UserID,
		AssignedTo:     req.AssignedTo,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Access:         req.Access,
		Title:          req.Title,
		Department:     req.Department,
		Email:          req.Email,
		AltEmail:       req.AltEmail,
		Phone:          req.Phone,
		Mobile:         req.Mobile,
		Fax:            req.Fax,
		DoNotCall:      req.DoNotCall,
		BackgroundInfo: req.BackgroundInfo,
	}

	if err := h.db.Create(&contact).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create contact")
		return
	}
	writeJSON(w, http.StatusCreated, contact)
}

func (h *WriteHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var contact model.Contact
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&contact).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, contact) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Access != "" {
		updates["access"] = req.Access
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Department != nil {
		updates["department"] = *req.Department
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.BackgroundInfo != nil {
		updates["background_info"] = *req.BackgroundInfo
	}

	if len(updates) > 0 {
		h.db.Model(&contact).Updates(updates)
	}

	h.db.First(&contact, id)
	writeJSON(w, http.StatusOK, contact)
}

func (h *WriteHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var contact model.Contact
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&contact).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, contact) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	h.db.Delete(&contact)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Opportunities ---

type opportunityRequest struct {
	Name           string   `json:"name"`
	AssignedTo     int64    `json:"assigned_to"`
	Access         string   `json:"access"`
	Source         *string  `json:"source,omitempty"`
	Stage          *string  `json:"stage,omitempty"`
	Probability    *int     `json:"probability,omitempty"`
	Amount         *float64 `json:"amount,omitempty"`
	Discount       *float64 `json:"discount,omitempty"`
	ClosesOn       *string  `json:"closes_on,omitempty"`
	CampaignID     *int64   `json:"campaign_id,omitempty"`
	BackgroundInfo *string  `json:"background_info,omitempty"`
}

func (h *WriteHandler) CreateOpportunity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req opportunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	if req.Access == "" {
		req.Access = model.AccessPublic
	}

	opp := model.Opportunity{
		UserID:         claims.UserID,
		AssignedTo:     req.AssignedTo,
		Name:           req.Name,
		Access:         req.Access,
		Source:         req.Source,
		Stage:          req.Stage,
		Probability:    req.Probability,
		Amount:         req.Amount,
		Discount:       req.Discount,
		CampaignID:     req.CampaignID,
		BackgroundInfo: req.BackgroundInfo,
	}
	if req.ClosesOn != nil {
		t, err := time.Parse("2006-01-02", *req.ClosesOn)
		if err == nil {
			opp.ClosesOn = &t
		}
	}

	if err := h.db.Create(&opp).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create opportunity")
		return
	}
	writeJSON(w, http.StatusCreated, opp)
}

func (h *WriteHandler) UpdateOpportunity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var opp model.Opportunity
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&opp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, opp) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req opportunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Access != "" {
		updates["access"] = req.Access
	}
	if req.Stage != nil {
		updates["stage"] = *req.Stage
	}
	if req.Probability != nil {
		updates["probability"] = *req.Probability
	}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Discount != nil {
		updates["discount"] = *req.Discount
	}
	if req.BackgroundInfo != nil {
		updates["background_info"] = *req.BackgroundInfo
	}

	if len(updates) > 0 {
		h.db.Model(&opp).Updates(updates)
	}

	h.db.First(&opp, id)
	writeJSON(w, http.StatusOK, opp)
}

func (h *WriteHandler) DeleteOpportunity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var opp model.Opportunity
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&opp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !entityWriteAccess(claims, opp) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	h.db.Delete(&opp)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

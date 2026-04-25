package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// convertLeadRequest is the JSON body for POST /leads/{id}/convert.
type convertLeadRequest struct {
	Account     convertAccountParams     `json:"account"`
	Opportunity convertOpportunityParams `json:"opportunity"`
	Access      string                   `json:"access"` // access level for new records
}

type convertAccountParams struct {
	ID   *int64 `json:"id,omitempty"`   // select existing account
	Name string `json:"name,omitempty"` // or create new by name
}

type convertOpportunityParams struct {
	Name        string   `json:"name"`
	Stage       *string  `json:"stage,omitempty"`
	Probability *int     `json:"probability,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	Discount    *float64 `json:"discount,omitempty"`
	ClosesOn    *string  `json:"closes_on,omitempty"`
}

// convertLeadResponse is returned on successful conversion.
type convertLeadResponse struct {
	Account     model.Account     `json:"account"`
	Contact     model.Contact     `json:"contact"`
	Opportunity model.Opportunity `json:"opportunity"`
}

// ConvertLead handles POST /leads/{id}/convert.
// Creates Account + Contact + Opportunity in a single transaction,
// links them via join tables, and marks the lead as "converted".
func (h *WriteHandler) ConvertLead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	leadID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// Fetch lead
	var lead model.Lead
	if err := h.db.Where("id = ? AND deleted_at IS NULL", leadID).First(&lead).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "lead not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Check write access
	if !entityWriteAccess(claims, lead) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Check if already converted
	if lead.Status != nil && *lead.Status == "converted" {
		writeError(w, http.StatusUnprocessableEntity, "lead is already converted")
		return
	}

	// Parse request
	var req convertLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Validate: must have account (by ID or name) and opportunity name
	if req.Account.ID == nil && req.Account.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "account id or name is required")
		return
	}
	if req.Opportunity.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "opportunity name is required")
		return
	}

	access := req.Access
	if access == "" {
		access = model.AccessPublic
	}

	var account model.Account
	var opp model.Opportunity
	var contact model.Contact

	txErr := h.db.Transaction(func(tx *gorm.DB) error {
		// 1. Find or create account
		if req.Account.ID != nil {
			if err := tx.Where("id = ? AND deleted_at IS NULL", *req.Account.ID).First(&account).Error; err != nil {
				return errors.New("account not found")
			}
		} else {
			// Try to find existing account by name first
			err := tx.Where("name = ? AND deleted_at IS NULL", req.Account.Name).First(&account).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new account
				account = model.Account{
					UserID:     claims.UserID,
					AssignedTo: claims.UserID,
					Name:       req.Account.Name,
					Access:     access,
				}
				if err := tx.Create(&account).Error; err != nil {
					return errors.New("failed to create account")
				}
			} else if err != nil {
				return errors.New("failed to lookup account")
			}
		}

		// 2. Create opportunity
		stage := "prospecting"
		if req.Opportunity.Stage != nil {
			stage = *req.Opportunity.Stage
		}
		opp = model.Opportunity{
			UserID:         claims.UserID,
			AssignedTo:     claims.UserID,
			Name:           req.Opportunity.Name,
			Access:         access,
			Stage:          &stage,
			Probability:    req.Opportunity.Probability,
			Amount:         req.Opportunity.Amount,
			Discount:       req.Opportunity.Discount,
			CampaignID:     lead.CampaignID,
			Source:         lead.Source,
		}
		if req.Opportunity.ClosesOn != nil {
			t, err := time.Parse("2006-01-02", *req.Opportunity.ClosesOn)
			if err == nil {
				opp.ClosesOn = &t
			}
		}
		if err := tx.Create(&opp).Error; err != nil {
			return errors.New("failed to create opportunity")
		}

		// 3. Create account_opportunity join
		if err := tx.Create(&model.AccountOpportunity{
			AccountID:     account.ID,
			OpportunityID: opp.ID,
		}).Error; err != nil {
			return errors.New("failed to link opportunity to account")
		}

		// 4. Create contact from lead data
		contact = model.Contact{
			UserID:         claims.UserID,
			LeadID:         &lead.ID,
			AssignedTo:     claims.UserID,
			FirstName:      lead.FirstName,
			LastName:       lead.LastName,
			Access:         access,
			Title:          lead.Title,
			Email:          lead.Email,
			AltEmail:       lead.AltEmail,
			Phone:          lead.Phone,
			Mobile:         lead.Mobile,
			DoNotCall:      lead.DoNotCall,
			BackgroundInfo: lead.BackgroundInfo,
		}
		if err := tx.Create(&contact).Error; err != nil {
			return errors.New("failed to create contact")
		}

		// 5. Create account_contact join
		if err := tx.Create(&model.AccountContact{
			AccountID: account.ID,
			ContactID: contact.ID,
		}).Error; err != nil {
			return errors.New("failed to link contact to account")
		}

		// 6. Create contact_opportunity join
		if err := tx.Create(&model.ContactOpportunity{
			ContactID:     contact.ID,
			OpportunityID: opp.ID,
		}).Error; err != nil {
			return errors.New("failed to link contact to opportunity")
		}

		// 7. Update counter caches
		tx.Exec("UPDATE accounts SET contacts_count = contacts_count + 1 WHERE id = ?", account.ID)
		tx.Exec("UPDATE accounts SET opportunities_count = opportunities_count + 1 WHERE id = ?", account.ID)
		if opp.CampaignID != nil {
			tx.Exec("UPDATE campaigns SET opportunities_count = opportunities_count + 1 WHERE id = ?", *opp.CampaignID)
		}

		// 8. Mark lead as converted
		tx.Model(&lead).Update("status", "converted")

		return nil
	})

	if txErr != nil {
		// Distinguish known validation errors from internal errors
		msg := txErr.Error()
		if msg == "account not found" {
			writeError(w, http.StatusNotFound, msg)
		} else {
			writeError(w, http.StatusUnprocessableEntity, msg)
		}
		return
	}

	// Record audit trail outside transaction (non-critical)
	h.versions.RecordCreate("Account", account.ID, claims.UserID, account)
	h.versions.RecordCreate("Opportunity", opp.ID, claims.UserID, opp)
	h.versions.RecordCreate("Contact", contact.ID, claims.UserID, contact)
	converted := "converted"
	h.versions.RecordUpdate("Lead", lead.ID, claims.UserID, lead, map[string]interface{}{"status": converted})

	writeJSON(w, http.StatusOK, convertLeadResponse{
		Account:     account,
		Contact:     contact,
		Opportunity: opp,
	})
}

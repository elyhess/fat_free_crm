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
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// WriteHandler provides create, update, delete endpoints for CRM entities.
type WriteHandler struct {
	db       *gorm.DB
	authzSvc *service.AuthorizationService
	versions *service.VersionRecorder
}

func NewWriteHandler(db *gorm.DB, authzSvc *service.AuthorizationService) *WriteHandler {
	return &WriteHandler{db: db, authzSvc: authzSvc, versions: service.NewVersionRecorder(db)}
}

// --- Tasks ---

type taskRequest struct {
	Name           string  `json:"name"`
	AssignedTo     int64   `json:"assigned_to"`
	Priority       *string `json:"priority,omitempty"`
	Category       *string `json:"category,omitempty"`
	Bucket         *string `json:"bucket,omitempty"`
	DueAt          *string `json:"due_at,omitempty"`
	AssetID        *int64  `json:"asset_id,omitempty"`
	AssetType      *string `json:"asset_type,omitempty"`
	BackgroundInfo *string `json:"background_info,omitempty"`
}

func (h *WriteHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	task := model.Task{
		UserID:         claims.UserID,
		AssignedTo:     req.AssignedTo,
		Name:           req.Name,
		Priority:       req.Priority,
		Category:       req.Category,
		Bucket:         req.Bucket,
		AssetID:        req.AssetID,
		AssetType:      req.AssetType,
		BackgroundInfo: req.BackgroundInfo,
	}
	if req.DueAt != nil {
		t, err := time.Parse(time.RFC3339, *req.DueAt)
		if err == nil {
			task.DueAt = &t
		}
	}

	if err := h.db.Create(&task).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *WriteHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
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

	var task model.Task
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Check ownership or admin
	if !claims.Admin && task.UserID != claims.UserID && task.AssignedTo != claims.UserID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AssignedTo != 0 {
		updates["assigned_to"] = req.AssignedTo
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Bucket != nil {
		updates["bucket"] = *req.Bucket
	}
	if req.DueAt != nil {
		t, err := time.Parse(time.RFC3339, *req.DueAt)
		if err == nil {
			updates["due_at"] = t
		}
	}
	if req.BackgroundInfo != nil {
		updates["background_info"] = *req.BackgroundInfo
	}

	if len(updates) > 0 {
		if err := h.db.Model(&task).Updates(updates).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update task")
			return
		}
	}

	h.db.First(&task, id)
	writeJSON(w, http.StatusOK, task)
}

func (h *WriteHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
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

	var task model.Task
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !claims.Admin && task.UserID != claims.UserID && task.AssignedTo != claims.UserID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.db.Delete(&task).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *WriteHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
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

	var task model.Task
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !claims.Admin && task.UserID != claims.UserID && task.AssignedTo != claims.UserID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	now := time.Now()
	h.db.Model(&task).Updates(map[string]interface{}{
		"completed_at": now,
		"completed_by": claims.UserID,
	})

	h.db.First(&task, id)
	writeJSON(w, http.StatusOK, task)
}

func (h *WriteHandler) UncompleteTask(w http.ResponseWriter, r *http.Request) {
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

	var task model.Task
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !claims.Admin && task.UserID != claims.UserID && task.AssignedTo != claims.UserID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	h.db.Model(&task).Updates(map[string]interface{}{
		"completed_at": nil,
		"completed_by": nil,
	})

	h.db.First(&task, id)
	writeJSON(w, http.StatusOK, task)
}

package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
)

// DashboardHandler serves aggregated dashboard data.
type DashboardHandler struct {
	db *gorm.DB
}

// NewDashboardHandler creates a new dashboard handler.
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// TaskBucket holds a bucket name and its tasks count.
type TaskBucket struct {
	Bucket string `json:"bucket"`
	Count  int64  `json:"count"`
}

// TaskSummaryResponse is the response for the task summary endpoint.
type TaskSummaryResponse struct {
	Buckets    []TaskBucket `json:"buckets"`
	TotalTasks int64        `json:"total_tasks"`
}

// TaskSummary returns pending tasks grouped by bucket for the current user.
// Matches Rails: Task.visible_on_dashboard(user) grouped by time buckets.
func (h *DashboardHandler) TaskSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	dayAfterTomorrow := todayStart.AddDate(0, 0, 2)

	// Calculate next week boundaries
	daysUntilMonday := (8 - int(todayStart.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	nextWeekStart := todayStart.AddDate(0, 0, daysUntilMonday)
	nextWeekEnd := nextWeekStart.AddDate(0, 0, 7)

	// Base query: pending tasks for this user (visible_on_dashboard logic)
	baseQuery := h.db.Table("tasks").
		Where("deleted_at IS NULL").
		Where("completed_at IS NULL").
		Where("(user_id = ? AND (assigned_to IS NULL OR assigned_to = 0)) OR assigned_to = ?", userID, userID)

	buckets := []struct {
		name  string
		query func(*gorm.DB) *gorm.DB
	}{
		{"due_asap", func(db *gorm.DB) *gorm.DB {
			return db.Where("due_at IS NULL AND bucket = 'due_asap'")
		}},
		{"overdue", func(db *gorm.DB) *gorm.DB {
			return db.Where("due_at < ?", todayStart)
		}},
		{"due_today", func(db *gorm.DB) *gorm.DB {
			return db.Where("due_at >= ? AND due_at < ?", todayStart, tomorrowStart)
		}},
		{"due_tomorrow", func(db *gorm.DB) *gorm.DB {
			return db.Where("due_at >= ? AND due_at < ?", tomorrowStart, dayAfterTomorrow)
		}},
		{"due_this_week", func(db *gorm.DB) *gorm.DB {
			return db.Where("due_at >= ? AND due_at < ?", dayAfterTomorrow, nextWeekStart)
		}},
		{"due_next_week", func(db *gorm.DB) *gorm.DB {
			return db.Where("due_at >= ? AND due_at < ?", nextWeekStart, nextWeekEnd)
		}},
		{"due_later", func(db *gorm.DB) *gorm.DB {
			return db.Where("(due_at IS NULL AND bucket = 'due_later') OR due_at >= ?", nextWeekEnd)
		}},
	}

	var resp TaskSummaryResponse
	for _, b := range buckets {
		var count int64
		q := baseQuery.Session(&gorm.Session{NewDB: true}).
			Table("tasks").
			Where("deleted_at IS NULL").
			Where("completed_at IS NULL").
			Where("(user_id = ? AND (assigned_to IS NULL OR assigned_to = 0)) OR assigned_to = ?", userID, userID)
		b.query(q).Count(&count)
		resp.Buckets = append(resp.Buckets, TaskBucket{Bucket: b.name, Count: count})
		resp.TotalTasks += count
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// PipelineStage holds an opportunity stage and its summary.
type PipelineStage struct {
	Stage       string  `json:"stage"`
	Count       int64   `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	WeightedSum float64 `json:"weighted_sum"`
}

// PipelineResponse is the response for the pipeline summary endpoint.
type PipelineResponse struct {
	Stages         []PipelineStage `json:"stages"`
	TotalCount     int64           `json:"total_count"`
	TotalAmount    float64         `json:"total_amount"`
	TotalWeighted  float64         `json:"total_weighted"`
}

// PipelineSummary returns opportunity pipeline data for the current user.
// Matches Rails: Opportunity.visible_on_dashboard(user).pipeline
func (h *DashboardHandler) PipelineSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	type stageRow struct {
		Stage       string  `json:"stage"`
		Count       int64   `json:"count"`
		TotalAmount float64 `json:"total_amount"`
		WeightedSum float64 `json:"weighted_sum"`
	}

	var rows []stageRow

	query := h.db.Table("opportunities").
		Select("stage, COUNT(*) as count, COALESCE(SUM(COALESCE(amount, 0) - COALESCE(discount, 0)), 0) as total_amount, COALESCE(SUM((COALESCE(amount, 0) - COALESCE(discount, 0)) * COALESCE(probability, 0) / 100.0), 0) as weighted_sum").
		Where("deleted_at IS NULL").
		Where("(stage IS NULL OR (stage != 'won' AND stage != 'lost'))").
		Where("(user_id = ? AND (assigned_to IS NULL OR assigned_to = 0)) OR assigned_to = ?", userID, userID).
		Group("stage").
		Order("stage")

	if err := query.Find(&rows).Error; err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := PipelineResponse{Stages: make([]PipelineStage, 0, len(rows))}
	for _, row := range rows {
		resp.Stages = append(resp.Stages, PipelineStage(row))
		resp.TotalCount += row.Count
		resp.TotalAmount += row.TotalAmount
		resp.TotalWeighted += row.WeightedSum
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

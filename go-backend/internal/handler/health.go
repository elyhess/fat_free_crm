package handler

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{Status: "ok", Database: "ok"}
	statusCode := http.StatusOK

	sqlDB, err := h.db.DB()
	if err != nil {
		resp.Status = "degraded"
		resp.Database = "error"
		statusCode = http.StatusServiceUnavailable
	} else if err := sqlDB.Ping(); err != nil {
		resp.Status = "degraded"
		resp.Database = "unreachable"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}

package handler

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
)

func NewRouter(db *gorm.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logging)

	health := NewHealthHandler(db)
	r.Get("/health", health.Health)

	return r
}

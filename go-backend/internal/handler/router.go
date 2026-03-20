package handler

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

func NewRouter(db *gorm.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logging)

	health := NewHealthHandler(db)
	r.Get("/health", health.Health)

	fieldGroupRepo := repository.NewFieldGroupRepository(db)
	fieldsSvc := service.NewCustomFieldService(fieldGroupRepo)
	fieldsHandler := NewFieldsHandler(fieldsSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/field_groups", fieldsHandler.ListFieldGroups)
	})

	return r
}

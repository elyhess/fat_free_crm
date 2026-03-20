package handler

import (
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// RouterConfig holds dependencies for router construction.
type RouterConfig struct {
	DB             *gorm.DB
	JWTSecret      string
	JWTExpiryHours int
}

func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logging)

	health := NewHealthHandler(cfg.DB)
	r.Get("/health", health.Health)

	jwtSvc := auth.NewJWTService(cfg.JWTSecret, time.Duration(cfg.JWTExpiryHours)*time.Hour)
	userRepo := repository.NewUserRepository(cfg.DB)
	authHandler := NewAuthHandler(userRepo, jwtSvc)

	fieldGroupRepo := repository.NewFieldGroupRepository(cfg.DB)
	fieldsSvc := service.NewCustomFieldService(fieldGroupRepo)
	fieldsHandler := NewFieldsHandler(fieldsSvc)

	authzSvc := service.NewAuthorizationService(cfg.DB)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/login", authHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtSvc))
			r.Get("/field_groups", fieldsHandler.ListFieldGroups)

			// Dashboard routes
			dashboard := NewDashboardHandler(cfg.DB)
			r.Get("/dashboard/tasks", dashboard.TaskSummary)
			r.Get("/dashboard/pipeline", dashboard.PipelineSummary)

			// Entity CRUD routes
			RegisterEntityRoutes(r, cfg.DB, authzSvc)

			// Supporting reads (comments, addresses, tags, versions, users)
			supportingRepo := repository.NewSupportingRepository(cfg.DB)
			supporting := NewSupportingHandler(supportingRepo)
			r.Get("/tags", supporting.ListAllTags)
			r.Get("/activity", supporting.ListRecentActivity)
			r.Get("/users", supporting.ListUsers)
			r.Get("/{entity}/{id}/comments", supporting.ListComments)
			r.Get("/{entity}/{id}/addresses", supporting.ListAddresses)
			r.Get("/{entity}/{id}/tags", supporting.ListEntityTags)
			r.Get("/{entity}/{id}/versions", supporting.ListEntityVersions)
		})
	})

	return r
}

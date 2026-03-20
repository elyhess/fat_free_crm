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

			// Write routes
			writes := NewWriteHandler(cfg.DB, authzSvc)

			// Tasks
			r.Post("/tasks", writes.CreateTask)
			r.Put("/tasks/{id}", writes.UpdateTask)
			r.Delete("/tasks/{id}", writes.DeleteTask)
			r.Put("/tasks/{id}/complete", writes.CompleteTask)
			r.Put("/tasks/{id}/uncomplete", writes.UncompleteTask)

			// Accounts
			r.Post("/accounts", writes.CreateAccount)
			r.Put("/accounts/{id}", writes.UpdateAccount)
			r.Delete("/accounts/{id}", writes.DeleteAccount)

			// Campaigns
			r.Post("/campaigns", writes.CreateCampaign)
			r.Put("/campaigns/{id}", writes.UpdateCampaign)
			r.Delete("/campaigns/{id}", writes.DeleteCampaign)

			// Leads
			r.Post("/leads", writes.CreateLead)
			r.Put("/leads/{id}", writes.UpdateLead)
			r.Delete("/leads/{id}", writes.DeleteLead)
			r.Put("/leads/{id}/reject", writes.RejectLead)

			// Contacts
			r.Post("/contacts", writes.CreateContact)
			r.Put("/contacts/{id}", writes.UpdateContact)
			r.Delete("/contacts/{id}", writes.DeleteContact)

			// Opportunities
			r.Post("/opportunities", writes.CreateOpportunity)
			r.Put("/opportunities/{id}", writes.UpdateOpportunity)
			r.Delete("/opportunities/{id}", writes.DeleteOpportunity)

			// Supporting writes (comments, tags, addresses)
			r.Post("/{entity}/{id}/comments", writes.CreateComment)
			r.Delete("/comments/{id}", writes.DeleteComment)
			r.Post("/{entity}/{id}/tags", writes.AddTag)
			r.Delete("/{entity}/{id}/tags/{tag_id}", writes.RemoveTag)
			r.Post("/{entity}/{id}/addresses", writes.CreateAddress)
			r.Delete("/addresses/{id}", writes.DeleteAddress)

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

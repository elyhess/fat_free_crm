package handler

import (
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/frontend"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// RouterConfig holds dependencies for router construction.
type RouterConfig struct {
	DB             *gorm.DB
	JWTSecret      string
	JWTExpiryHours int
	AvatarDir      string // base directory for avatar file storage (defaults to ".")
	ServeFrontend  bool   // serve embedded React SPA for non-API routes
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

		// Auth flows (public — no JWT required)
		emailSvc := service.NewEmailService(service.EmailConfig{
			Host:     envOrDefault("SMTP_HOST", "localhost"),
			Port:     envOrDefault("SMTP_PORT", "25"),
			Username: envOrDefault("SMTP_USERNAME", ""),
			Password: envOrDefault("SMTP_PASSWORD", ""),
			From:     envOrDefault("SMTP_FROM", "noreply@fatfreecrm.com"),
		})
		baseURL := envOrDefault("FRONTEND_URL", "http://localhost:3000")
		authFlows := NewAuthFlowsHandler(cfg.DB, emailSvc, baseURL)
		r.Post("/auth/forgot-password", authFlows.ForgotPassword)
		r.Post("/auth/reset-password", authFlows.ResetPassword)
		r.Post("/auth/register", authFlows.Register)
		r.Post("/auth/confirm", authFlows.ConfirmEmail)
		r.Post("/auth/resend-confirmation", authFlows.ResendConfirmation)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtSvc))
			r.Get("/field_groups", fieldsHandler.ListFieldGroups)

			// Dashboard routes
			dashboard := NewDashboardHandler(cfg.DB)
			r.Get("/dashboard/tasks", dashboard.TaskSummary)
			r.Get("/dashboard/pipeline", dashboard.PipelineSummary)

			// Profile (user self-service)
			profile := NewProfileHandler(cfg.DB)
			r.Get("/profile", profile.GetProfile)
			r.Put("/profile", profile.UpdateProfile)
			r.Put("/profile/password", profile.ChangePassword)

			// Avatar
			avatarDir := "."
			if cfg.AvatarDir != "" {
				avatarDir = cfg.AvatarDir
			}
			avatar := NewAvatarHandler(cfg.DB, avatarDir)
			r.Post("/profile/avatar", avatar.UploadAvatar)
			r.Delete("/profile/avatar", avatar.DeleteAvatar)
			r.Get("/avatars/{user_id}", avatar.ServeAvatar)

			// Search
			search := NewSearchHandler(cfg.DB, authzSvc)
			r.Get("/search", search.Search)

			// Saved searches
			savedSearches := NewSavedSearchHandler(cfg.DB)
			r.Get("/saved_searches", savedSearches.List)
			r.Post("/saved_searches", savedSearches.Create)
			r.Put("/saved_searches/{id}", savedSearches.Update)
			r.Delete("/saved_searches/{id}", savedSearches.Delete)

			// Entity CRUD routes
			RegisterEntityRoutes(r, cfg.DB, authzSvc)

			// Entity relationship routes
			rels := NewRelationshipHandler(cfg.DB, authzSvc)
			r.Get("/accounts/{id}/contacts", rels.AccountContacts)
			r.Get("/accounts/{id}/opportunities", rels.AccountOpportunities)
			r.Get("/campaigns/{id}/leads", rels.CampaignLeads)
			r.Get("/campaigns/{id}/opportunities", rels.CampaignOpportunities)
			r.Get("/contacts/{id}/opportunities", rels.ContactOpportunities)

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
			r.Post("/leads/{id}/convert", writes.ConvertLead)

			// Contacts
			r.Post("/contacts", writes.CreateContact)
			r.Put("/contacts/{id}", writes.UpdateContact)
			r.Delete("/contacts/{id}", writes.DeleteContact)

			// Opportunities
			r.Post("/opportunities", writes.CreateOpportunity)
			r.Put("/opportunities/{id}", writes.UpdateOpportunity)
			r.Delete("/opportunities/{id}", writes.DeleteOpportunity)

			// Subscription routes
			subs := NewSubscriptionHandler(cfg.DB)
			r.Post("/{entity}/{id}/subscribe", subs.Subscribe)
			r.Post("/{entity}/{id}/unsubscribe", subs.Unsubscribe)
			r.Get("/{entity}/{id}/subscription", subs.GetSubscription)

			// Supporting writes (comments, tags, addresses)
			r.Post("/{entity}/{id}/comments", writes.CreateComment)
			r.Delete("/comments/{id}", writes.DeleteComment)
			r.Post("/{entity}/{id}/tags", writes.AddTag)
			r.Delete("/{entity}/{id}/tags/{tag_id}", writes.RemoveTag)
			r.Post("/{entity}/{id}/addresses", writes.CreateAddress)
			r.Delete("/addresses/{id}", writes.DeleteAddress)

			// Admin routes (admin-only, enforced by handler)
			admin := NewAdminHandler(cfg.DB)
			r.Post("/admin/users", admin.CreateUser)
			r.Put("/admin/users/{id}", admin.UpdateUser)
			r.Delete("/admin/users/{id}", admin.DeleteUser)
			r.Put("/admin/users/{id}/suspend", admin.SuspendUser)
			r.Put("/admin/users/{id}/reactivate", admin.ReactivateUser)
			r.Get("/admin/groups", admin.ListGroups)
			r.Post("/admin/groups", admin.CreateGroup)
			r.Put("/admin/groups/{id}", admin.UpdateGroup)
			r.Delete("/admin/groups/{id}", admin.DeleteGroup)
			r.Post("/admin/field_groups", admin.CreateFieldGroup)
			r.Put("/admin/field_groups/{id}", admin.UpdateFieldGroup)
			r.Delete("/admin/field_groups/{id}", admin.DeleteFieldGroup)

			// Admin settings
			settings := NewSettingsHandler(cfg.DB)
			r.Get("/admin/settings", settings.GetSettings)
			r.Put("/admin/settings", settings.UpdateSettings)

			// Admin extras (plugins, research tools)
			extras := NewAdminExtrasHandler(cfg.DB)
			r.Get("/admin/plugins", extras.ListPlugins)
			r.Get("/admin/research_tools", extras.ListResearchTools)
			r.Post("/admin/research_tools", extras.CreateResearchTool)
			r.Put("/admin/research_tools/{id}", extras.UpdateResearchTool)
			r.Delete("/admin/research_tools/{id}", extras.DeleteResearchTool)

			// Admin field CRUD
			adminFields := NewAdminFieldsHandler(cfg.DB, fieldsSvc)
			r.Post("/admin/fields", adminFields.CreateField)
			r.Put("/admin/fields/{id}", adminFields.UpdateField)
			r.Delete("/admin/fields/{id}", adminFields.DeleteField)
			r.Post("/admin/fields/sort", adminFields.SortFields)

			// Custom field values on entities
			r.Get("/{entity}/{id}/custom_fields", adminFields.GetEntityCustomFields)
			r.Put("/{entity}/{id}/custom_fields", adminFields.UpdateEntityCustomFields)

			// Autocomplete routes
			ac := NewAutocompleteHandler(cfg.DB, authzSvc)
			r.Get("/accounts/autocomplete", ac.Accounts)
			r.Get("/contacts/autocomplete", ac.Contacts)
			r.Get("/leads/autocomplete", ac.Leads)
			r.Get("/campaigns/autocomplete", ac.Campaigns)
			r.Get("/opportunities/autocomplete", ac.Opportunities)

			// Export/Import routes
			export := NewExportHandler(cfg.DB, authzSvc)
			r.Get("/accounts/export", export.ExportAccounts)
			r.Get("/contacts/export", export.ExportContacts)
			r.Get("/leads/export", export.ExportLeads)
			r.Get("/opportunities/export", export.ExportOpportunities)
			r.Get("/campaigns/export", export.ExportCampaigns)
			r.Get("/tasks/export", export.ExportTasks)
			imp := NewImportHandler(cfg.DB)
			r.Post("/accounts/import", imp.ImportAccounts)
			r.Post("/contacts/import", imp.ImportContacts)
			r.Post("/leads/import", imp.ImportLeads)
			r.Get("/{entity}/import/template", imp.ImportTemplate)
			r.Get("/contacts/export/vcard", imp.VCardExportContacts)

			// Email routes
			emails := NewEmailHandler(cfg.DB)
			r.Get("/{entity}/{id}/emails", emails.ListEmails)
			r.Delete("/emails/{id}", emails.DeleteEmail)

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

	// Serve embedded React SPA for non-API routes
	if cfg.ServeFrontend {
		r.NotFound(frontend.Handler().ServeHTTP)
	}

	return r
}

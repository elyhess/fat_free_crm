# Remaining Work — 2026-03-21

Status snapshot of what's done and what still needs to be built to reach full feature parity with the Rails app.

## What's Done

- All 6 entities: full CRUD (list, detail, create, update, delete)
- CSV export for all entities, CSV import for accounts/contacts/leads
- vCard export for contacts
- Comments, tags, addresses (polymorphic CRUD)
- Authentication (JWT, Devise-compatible password hashing)
- Authorization (Public/Private/Shared, admin bypass, group membership)
- Audit trail (version records on create/update/destroy for tracked entities)
- Cross-entity search (LIKE-based, authorization-scoped)
- Dashboard (task summary by bucket, pipeline by stage)
- Admin: user management (CRUD, suspend, reactivate), group management, field group management
- User self-service (profile edit, password change)
- React frontend: login, dashboard, all entity list/detail pages, create/edit/delete, search, profile

## What's Left

### Priority 1 — Required for Full Parity

#### ~~Lead Conversion~~ (DONE)
- ~~**Go:** `POST /leads/{id}/convert`~~
- ~~**React:** Conversion form on lead detail page~~
- Completed: Transactional endpoint creating Account + Contact + Opportunity with join tables, counter caches, audit trail. React form with account selector and opportunity fields. 11 Go tests.

#### ~~Custom Field System (End-to-End)~~ (DONE)
- ~~**Go:** Read dynamic `cf_*` columns from entity tables and include in API responses~~
- ~~**Go:** Write custom field values on entity create/update~~
- ~~**Go:** Admin CRUD for field definitions (creates/drops columns on entity tables via `ALTER TABLE`)~~
- ~~**React:** Render custom fields in entity forms (load field definitions, dynamic form generation)~~
- ~~**React:** Admin UI for managing field definitions~~
- Completed: Go endpoints for reading/writing cf_* values (GET/PUT /{entity}/{id}/custom_fields), admin field CRUD (POST/PUT/DELETE /admin/fields, POST /admin/fields/sort) with ALTER TABLE column management, safe type transitions, collision-safe naming. React: CustomFieldsDisplay on all 6 entity detail pages, CustomFieldsForm for editing, AdminFieldsPage with tabbed entity view. 20+ Go tests.

#### ~~Application Settings~~ (DONE)
- ~~**Go:** `GET /admin/settings`, `PUT /admin/settings` — app-wide configuration (company name, base URL, email config, etc.)~~
- ~~**React:** Admin settings page~~
- Completed: Go endpoints with YAML deserialization/serialization, admin-only auth, bulk update in transaction. React admin page with grouped sections (General, Options, Dropdowns, SMTP, Email Dropbox). 10 Go tests.

#### ~~Entity Relationship Endpoints~~ (DONE)
- ~~**Go:** Endpoints to list related entities~~
- ~~**React:** Show related entities on detail pages~~
- Completed: 5 Go endpoints with auth scoping + pagination, RelatedEntities React component on Account/Campaign/Contact detail pages

#### ~~Subscriptions / Notifications~~ (DONE)
- ~~**Go:** Subscribe/unsubscribe endpoints~~
- ~~**React:** Subscribe button on entity detail pages~~
- Completed: 3 Go endpoints (subscribe, unsubscribe, get state) with YAML-compatible serialization for all 6 entity types. SubscribeButton component in sidebar. 9 Go tests.

### Priority 2 — Important but Not Blocking

#### ~~Dashboard Activity Feed~~ (DONE)
- ~~**Go:** `GET /dashboard/activity` — recent versions across all tracked entities~~
- ~~**React:** Activity timeline on dashboard page~~
- Completed: Activity feed on dashboard using existing GET /activity endpoint. Color-coded events (create/update/destroy), entity links, relative timestamps.

#### ~~User Avatar~~ (DONE)
- ~~**Go:** `POST /profile/avatar` — file upload, `DELETE /profile/avatar`, `GET /avatars/{user_id}` — serve with Gravatar fallback~~
- ~~**React:** Avatar in navbar + profile page with upload/remove~~
- Completed: Multipart upload (PNG/JPEG/GIF, max 5MB), disk storage in uploads/avatars/, content-type detection from file bytes, Gravatar redirect fallback. React: avatar in navbar, profile page with upload/remove buttons. 11 Go tests.

#### ~~Autocomplete Endpoints~~ (DONE)
- ~~**Go:** Typeahead search for entity selection (used in forms — e.g. pick an account when creating a contact)~~
- ~~`GET /{entity}/autocomplete?q=...` returning `[{id, name}]` for all 5 entities~~
- ~~**React:** EntityAutocomplete component with debounced search, integrated into contact and opportunity forms~~
- Completed: 5 Go endpoints with ILIKE search, auth scoping, limit 10. React EntityAutocomplete component with debounced typeahead, dropdown, clear button. Integrated as 'autocomplete' field type in EntityForm. 11 Go tests.

#### ~~Attachment System~~ (N/A — Not a Rails Feature)
- Rails' `attach`/`discard` actions link entities to each other (e.g. Tasks to Accounts), not file uploads. ActiveStorage is only used for avatars.
- Entity relationships are already covered by RelatedEntities endpoints and React component.
- Avatars are covered by the User Avatar feature above.
- No file attachment feature exists in Rails to port.

### Priority 3 — Deferred / Post-MVP

#### ~~Email Integration (Phase 3.1)~~ (DONE)
- ~~IMAP email fetching (dropbox — attach emails to CRM entities)~~
- ~~Email reply parsing (comment replies via email)~~
- ~~Email sending (SMTP via go-mail)~~
- ~~Inline CSS for email templates~~
- Completed: Email model with polymorphic mediator (Account/Contact/Lead/Opportunity/Campaign). REST endpoints for listing entity emails and soft-delete. IMAP dropbox processor with entity matching by first-line keyword and recipient email address. Comment reply processor parsing [entity_type:id] from subjects with quoted-content stripping. Notification mailers for assignments, comments, and dropbox attachments. Goroutine-based background scheduler. React EmailTimeline component on all entity detail pages. 11 Go tests (4 handler, 7 service).

#### ~~Background Jobs (Phase 3.3)~~ (DONE)
- ~~Job queue (River or Asynq)~~
- ~~Email processing jobs (IMAP polling)~~
- ~~Any deferred/async work~~
- Completed: Simple goroutine-based scheduler with periodic job execution, used for IMAP dropbox and comment reply polling.

#### ~~Advanced Search & Filtering~~ (DONE)
- ~~Full-text search upgrade (Postgres `tsvector`)~~
- ~~Advanced filtering UI in React (replaces Ransack)~~
- ~~Saved searches / views~~
- Completed: PostgreSQL tsvector columns + GIN indexes on 5 entity tables with auto-update triggers. Search handler uses ts_rank for relevance-ranked results. Entity list filtering via `filter[field_op]=value` query params with column allowlist (eq, cont, gt, lt, blank, present operators). React: entity type dropdown on search page, filter bars on Accounts/Leads/Opportunities/Campaigns. Saved searches CRUD scoped per user with JSONB filters. 13 Go tests.

#### UI Enhancements
- Inline editing on entity list/detail pages
- Opportunity stage pipeline drag-and-drop
- Dashboard UI state (toggle views, timezone, timeline options)

#### ~~Auth Flows~~ (DONE)
- ~~Password reset via email (Devise `passwords_controller`)~~
- ~~Email confirmation flow (Devise `confirmations_controller`)~~
- ~~User registration (Devise `registrations_controller`)~~
- Completed: Go endpoints for forgot-password, reset-password, register, confirm, resend-confirmation. SHA-256 hashed tokens stored in DB, 6-hour reset expiry, signup setting (allowed/needs_approval/not_allowed), email enumeration prevention. SMTP email service with password reset, confirmation, and welcome templates. React pages for forgot password, reset password, registration, email confirmation. Login page links. 12 Go tests.

#### ~~Admin Extras~~ (DONE)
- ~~Plugin listing (read-only — `GET /admin/plugins`)~~
- ~~Research tools CRUD (admin-configurable lookup tools)~~
- ~~Field sorting/reordering~~
- Completed: Plugin listing returns empty list (Rails feature is a stub). Research tools full CRUD with admin-only auth. React admin page with table, inline create/edit forms, enable/disable toggle. Field sorting was already implemented. 7 Go tests.

### ~~Database Migration Ownership~~ (DONE)

- Completed: goose-based migration system in `cmd/migrate` with baseline migration from `pg_dump`. Supports up/down/status/create/mark-applied. Baseline tested on fresh database. Makefile targets added. Documented in `docs/local-development.md`.
- `goose_db_version` table lives alongside Rails' `schema_migrations` — no conflict.
- New migrations go in `go-backend/db/migrations/` as numbered SQL files.

### Phase 4 — Decommission Rails

Once all Priority 1 and Priority 2 items are complete:

- [ ] Audit: compare every Rails route to a Go equivalent
- [ ] Data verification: confirm Go reads/writes produce identical results to Rails
- [ ] Remove Rails proxy/routing — Go serves all traffic
- [ ] Serve React static assets from Go (embed in binary or reverse proxy)
- [ ] Clean up Rails-era schema artifacts
- [ ] Update deployment configuration
- [ ] Update CLAUDE.md and documentation

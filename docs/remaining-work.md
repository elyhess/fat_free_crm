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

#### Lead Conversion
- **Go:** `POST /leads/{id}/convert` — creates an Account + Contact + Opportunity in a single transaction, marks lead as converted, decrements campaign counter
- **React:** Conversion form on lead detail page (select/create account, fill opportunity fields)
- **Complexity:** High — multi-entity transactional write with campaign counter management
- **Reference:** `app/controllers/entities/leads_controller.rb` (`convert` and `promote` actions)

#### Custom Field System (End-to-End)
- **Go:** Read dynamic `cf_*` columns from entity tables and include in API responses
- **Go:** Write custom field values on entity create/update
- **Go:** Admin CRUD for field definitions (creates/drops columns on entity tables via `ALTER TABLE`)
- **React:** Render custom fields in entity forms (load field definitions, dynamic form generation)
- **React:** Admin UI for managing field definitions
- **Complexity:** High — dynamic schema modification, varies by field type (string, integer, date, etc.)
- **Reference:** `app/models/fields/`, `lib/fat_free_crm/custom_fields.rb`, `app/controllers/admin/fields_controller.rb`

#### Application Settings
- **Go:** `GET /admin/settings`, `PUT /admin/settings` — app-wide configuration (company name, base URL, email config, etc.)
- **React:** Admin settings page
- **Reference:** `app/controllers/admin/settings_controller.rb`, `app/models/setting.rb`

#### Entity Relationship Endpoints
- **Go:** Endpoints to list related entities:
  - `GET /accounts/{id}/contacts`
  - `GET /accounts/{id}/opportunities`
  - `GET /campaigns/{id}/leads`
  - `GET /campaigns/{id}/opportunities`
  - `GET /contacts/{id}/opportunities`
- **React:** Show related entities on detail pages (e.g. contacts belonging to an account)
- **Complexity:** Low — filtered queries using existing join tables

#### Subscriptions / Notifications
- **Go:** `POST /{entity}/{id}/subscribe`, `POST /{entity}/{id}/unsubscribe`
- Manages `subscribed_users` field on entities — controls who gets notified about changes
- **React:** Subscribe/unsubscribe button on entity detail pages
- **Reference:** `app/models/users/user.rb` (subscription methods)

### Priority 2 — Important but Not Blocking

#### Dashboard Activity Feed
- **Go:** `GET /dashboard/activity` — recent versions across all tracked entities, scoped to current user's accessible records
- **React:** Activity timeline on dashboard page
- **Note:** Version reads already exist (`GET /activity`), but the dashboard React component doesn't display it yet

#### User Avatar
- **Go:** `POST /profile/avatar` — file upload for user profile picture
- **Go:** Serve avatar images
- **React:** Avatar display in nav + profile page, upload form
- **Complexity:** Medium — needs file storage (local disk or S3)

#### Autocomplete Endpoints
- **Go:** Typeahead search for entity selection (used in forms — e.g. pick an account when creating a contact)
- Likely: `GET /accounts/autocomplete?q=...` returning `[{id, name}]`
- **React:** Autocomplete/select components in entity forms

#### Attachment System
- **Go:** `PUT /{entity}/{id}/attach` — attach files to CRM records
- **Go:** File storage + retrieval
- **React:** File upload UI on entity detail pages, attachment list
- **Complexity:** High — file storage, MIME types, size limits
- **Reference:** Rails uses ActiveStorage

### Priority 3 — Deferred / Post-MVP

#### Email Integration (Phase 3.1)
- IMAP email fetching (dropbox — attach emails to CRM entities)
- Email reply parsing (comment replies via email)
- Email sending (SMTP via go-mail)
- Inline CSS for email templates
- **Complexity:** Very high — requires background jobs, IMAP connectivity

#### Background Jobs (Phase 3.3)
- Job queue (River or Asynq)
- Email processing jobs (IMAP polling)
- Any deferred/async work
- **Note:** Only needed once email integration is in scope

#### Advanced Search & Filtering
- Full-text search upgrade (Postgres `tsvector`)
- Advanced filtering UI in React (replaces Ransack)
- Saved searches / views

#### UI Enhancements
- Inline editing on entity list/detail pages
- Opportunity stage pipeline drag-and-drop
- Dashboard UI state (toggle views, timezone, timeline options)

#### Auth Flows
- Password reset via email (Devise `passwords_controller`)
- Email confirmation flow (Devise `confirmations_controller`)
- User registration (Devise `registrations_controller`)
- **Note:** Currently users are created by admin only

#### Admin Extras
- Plugin listing (read-only — `GET /admin/plugins`)
- Research tools CRUD (admin-configurable lookup tools)
- Field sorting/reordering

### Phase 4 — Decommission Rails

Once all Priority 1 and Priority 2 items are complete:

- [ ] Audit: compare every Rails route to a Go equivalent
- [ ] Data verification: confirm Go reads/writes produce identical results to Rails
- [ ] Remove Rails proxy/routing — Go serves all traffic
- [ ] Serve React static assets from Go (embed in binary or reverse proxy)
- [ ] Clean up Rails-era schema artifacts
- [ ] Update deployment configuration
- [ ] Update CLAUDE.md and documentation

# Go + React Migration Plan

A phased plan for migrating Fat Free CRM from Rails to a Go backend with a React frontend, using the strangler fig pattern.

## Guiding Principles

- **No big-bang rewrite.** Both systems run in parallel; Rails keeps serving traffic while Go takes over piece by piece.
- **Same database.** The Go service reads and writes the same Postgres database as Rails. No data migration needed until Rails is fully decommissioned.
- **One entity or concern per milestone.** Each piece is self-contained, testable, and demo-able before moving to the next.
- **Custom fields first.** The dynamic field system is the architectural linchpin — validate the Go approach before building on top of it.

---

## Phase 0: Foundation

Goal: Scaffold the Go project and solve the hardest architectural problem up front.

### 0.1 — Project Setup
- [x] Initialize Go module and repo structure
- [x] Choose and configure router (chi)
- [x] Set up database connection (pgx + GORM)
- [x] Configuration management (env vars, config files)
- [x] Docker Compose for local dev (Go service + Postgres)
- [x] CI pipeline (linting with golangci-lint, tests)
- [x] Logging and error handling patterns

### 0.2 — Custom Fields System
- [x] Design approach: reuse existing `fields` / `field_groups` tables (no JSONB migration needed)
- [x] Go models mapping Rails schema for field_groups and fields
- [x] Repository + service layers to read field definitions and validate custom data
- [x] API endpoint: GET /api/v1/field_groups?entity=Account
- [x] Document the approach (see docs/checklists/phase-0-2-custom-fields.md)
- [ ] Dynamic cf_* column reading from entity tables (deferred to entity read phase)
- [ ] Paired date range validation (deferred — no paired fields in DB yet)

### 0.3 — Authentication
- [x] Implement JWT-based auth in Go (HS256, configurable expiry)
- [x] Password verification compatible with existing Devise-encrypted passwords (authlogic_sha512)
- [x] Login endpoint (POST /api/v1/auth/login) — accepts username or email
- [x] Password complexity rules (matching devise-security config)
- [x] Middleware for protected routes (Bearer token)
- [x] User status checks (confirmed, not suspended)

### 0.4 — Authorization
- [x] Permission and Group models mapping Rails schema
- [x] Access control service (Public/Private/Shared logic — no Casbin, custom implementation)
- [x] Admin bypass (can manage all)
- [x] Owner/assignee check for Private records
- [x] Shared record check via permissions table (user + group)
- [x] Query scope builder (ScopeAccessible) for filtered entity lists
- [ ] Authorization middleware (deferred — will wire into router in Phase 1)
- [x] Tests (13 tests covering all access scenarios)

---

## Phase 1: Read-Only API

Goal: Go serves all read endpoints. React frontend consumes them. Rails still handles writes.

### 1.1 — React Frontend Scaffold
- [x] Initialize React project (Vite + TypeScript + Tailwind CSS)
- [x] Set up routing (React Router v7)
- [x] Auth flow (login page, token storage, protected routes)
- [x] Layout shell (nav, sidebar, dashboard skeleton)
- [x] API client layer (fetch wrapper with auth headers + 401 redirect)

### 1.2 — Dashboard
- [ ] Go endpoint: activity feed / recent items (deferred — requires versions table)
- [x] Go endpoint: task summary (grouped by bucket — overdue, today, tomorrow, etc.)
- [x] Go endpoint: pipeline/opportunity summary (stages, amounts, weighted values)
- [ ] React: dashboard page (wired to live endpoints)

### 1.3 — Entities (Read)

Migrate reads one entity at a time. For each entity:
- [x] Go: list endpoint with pagination, sorting, filtering
- [x] Go: detail endpoint with authorization scope
- [ ] Go: search endpoint (replaces Ransack)
- [ ] React: list view
- [ ] React: detail view

Order (simplest → most complex):
1. [x] Tasks (list + detail)
2. [x] Campaigns (list + detail)
3. [x] Leads (list + detail)
4. [x] Accounts (list + detail)
5. [x] Contacts (list + detail)
6. [x] Opportunities (list + detail)

### 1.4 — Supporting Reads
- [x] Comments (polymorphic — list per entity)
- [x] Tags (list all, list per entity via taggings join)
- [x] Audit log / versions (per entity + recent activity feed)
- [x] Users (admin-only, sensitive fields stripped)
- [x] Addresses (polymorphic, soft-delete filtered)

---

## Phase 2: Write Paths

Goal: Go handles all CRUD. Rails is no longer needed for day-to-day operations.

For each entity, migrate in this order:
1. Create
2. Update
3. Delete
4. Association management (link/unlink related records)

### 2.1 — Entity Writes

Same order as reads:
- [x] Tasks (create, update, complete/uncomplete, delete)
- [x] Campaigns (create, update, delete)
- [x] Leads (create, update, reject, delete + campaign counter management)
- [ ] Leads convert (promote to Account + Contact + Opportunity — deferred)
- [x] Accounts (create, update, delete)
- [x] Contacts (create, update, delete)
- [x] Opportunities (create, update with stage transitions, delete)

### 2.2 — Supporting Writes
- [x] Comments (add, delete on any entity)
- [x] Tags (add, remove)
- [x] Addresses (add, delete)
- [ ] Audit trail generation (Go writes version records on every mutation)
- [ ] Custom field values (create, update per entity)

### 2.3 — Admin Functions
- [ ] User management (create, suspend, activate, promote/demote admin)
- [ ] Group management
- [ ] Field group / custom field definition management
- [ ] Application settings

### 2.4 — React Write UIs
- [ ] Forms for each entity (with custom field rendering from definitions)
- [ ] Inline editing
- [ ] Delete confirmations
- [ ] Lead conversion flow
- [ ] Opportunity stage pipeline (drag-and-drop or stage selector)

---

## Phase 3: Supporting Systems

Goal: Migrate everything that isn't core CRUD.

### 3.1 — Email Integration
- [ ] IMAP email fetching (dropbox — attach emails to CRM entities)
- [ ] Email reply parsing (comment replies via email)
- [ ] Email sending (SMTP via go-mail)
- [ ] Inline CSS for email templates (go-premailer)

### 3.2 — Import / Export
- [ ] CSV export for all entities
- [ ] CSV import with field mapping
- [ ] vCard import/export for contacts

### 3.3 — Background Jobs
- [ ] Set up River (Postgres-backed job queue) or Asynq (Redis-backed)
- [ ] Email processing jobs (IMAP polling)
- [ ] Any deferred/async work

### 3.4 — Search
- [ ] Full-text search (Postgres `tsvector` or Elasticsearch if needed)
- [ ] Advanced filtering UI in React (replaces Ransack UI)
- [ ] Saved searches / views

---

## Phase 4: Decommission Rails

Goal: Rails is fully shut down. Go + React handles everything.

- [ ] Audit: compare feature parity checklist (every Rails route has a Go equivalent)
- [ ] Data verification: confirm Go reads/writes produce identical results
- [ ] Remove Rails proxy/routing — Go serves all traffic
- [ ] Clean up any Rails-era schema artifacts (if needed)
- [ ] Update deployment (single Go binary + React static assets)
- [ ] Update documentation and CLAUDE.md

---

## Per-Entity Migration Checklist Template

Use this for each entity (copy and fill in):

```
### [Entity Name]

**Read**
- [ ] Go: list endpoint (pagination, sort, filter)
- [ ] Go: detail endpoint (with associations)
- [ ] Go: search
- [ ] React: list page
- [ ] React: detail page
- [ ] Tests: Go handler + service tests
- [ ] Tests: React component tests

**Write**
- [ ] Go: create endpoint
- [ ] Go: update endpoint
- [ ] Go: delete endpoint
- [ ] Go: association management
- [ ] React: create/edit form (with custom fields)
- [ ] React: delete flow
- [ ] Tests: Go write path tests
- [ ] Tests: React form tests
- [ ] Audit trail: version records created on mutations

**Verified**
- [ ] Data parity confirmed (Go reads what Rails wrote, and vice versa)
- [ ] Custom fields working
- [ ] Authorization rules enforced
```

---

## Session Workflow

Each working session follows this pattern:

1. **Pick the next unchecked item** from the plan above.
2. **Read the existing Rails code** — model, controller, views, and specs for that piece.
3. **Build the Go equivalent** — handler, service, repository layers.
4. **Write tests.**
5. **Build the React component** (if applicable).
6. **Verify parity** — compare behavior with the Rails version.
7. **Check off the item** and commit.

---

## Reference Documents

- [Go Dependency Mapping](go_dependency_mapping.md) — Ruby gem → Go library equivalents
- `CLAUDE.md` — Project setup, architecture, and credentials

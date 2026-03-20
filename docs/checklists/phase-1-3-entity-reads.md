# Phase 1.3 — Entity Read Endpoints

## Checklist

### Go Models
- [x] Account model (maps Rails schema)
- [x] Contact model
- [x] Lead model
- [x] Opportunity model (with WeightedAmount)
- [x] Campaign model
- [x] Task model (with bucket constants)

### Generic Infrastructure
- [x] PaginationParams and PaginatedResult types
- [x] Generic EntityRepository with List and FindByID
- [x] SQL injection prevention (allowedSortColumns whitelist)
- [x] Soft delete filtering (deleted_at IS NULL)
- [x] Authorization scope integration (ScopeAccessible)

### API Endpoints (all behind JWT auth)
- [x] GET /api/v1/accounts (list + pagination)
- [x] GET /api/v1/accounts/{id} (detail)
- [x] GET /api/v1/contacts (list + pagination)
- [x] GET /api/v1/contacts/{id} (detail)
- [x] GET /api/v1/leads (list + pagination)
- [x] GET /api/v1/leads/{id} (detail)
- [x] GET /api/v1/opportunities (list + pagination)
- [x] GET /api/v1/opportunities/{id} (detail)
- [x] GET /api/v1/campaigns (list + pagination)
- [x] GET /api/v1/campaigns/{id} (detail)
- [x] GET /api/v1/tasks (list + pagination)
- [x] GET /api/v1/tasks/{id} (detail)

### Tests
- [x] List with empty data
- [x] List with data
- [x] Pagination (page, per_page, total_pages)
- [x] Access control filtering (Public vs Private)
- [x] Detail found
- [x] Detail not found (404)
- [x] Detail access denied (filtered to 404)
- [x] Soft-deleted records excluded
- [x] No token returns 401
- [x] All 6 entity types tested

### Deferred
- [ ] Search/filter endpoints (Phase 1.3 search — replaces Ransack)
- [ ] Association loading (contacts with account, etc.)
- [ ] React list and detail views

## Design Decisions

- **Generic repository**: Single `EntityRepository[T]` handles all entities, avoiding boilerplate.
- **Sort column whitelist**: Prevents SQL injection by only allowing known column names.
- **Scope-based authorization**: Access control applied as GORM scopes in the query, not post-fetch filtering.
- **404 for unauthorized**: Unauthorized detail requests return 404 (not 403) to avoid leaking record existence.

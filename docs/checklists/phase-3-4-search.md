# Phase 3.4 — Search

## Checklist

### Go Backend
- [x] GET /api/v1/search?q=term (cross-entity search)
- [x] Optional entity filter (?entity=accounts)
- [x] Authorization-scoped results (admin sees all, non-admin sees accessible)
- [x] LIKE-based search across name/email/company fields per entity
- [x] Results grouped by entity type with total count
- [x] 25 results per entity type limit

### React Frontend
- [x] Search bar in Layout nav (form submit navigates to /search?q=term)
- [x] SearchPage with grouped results by entity type
- [x] Result count per section
- [x] Empty state for no results
- [x] Route: /search

### Tests (5 tests)
- [x] Search across all entities (matches in accounts, contacts, opportunities, campaigns, tasks)
- [x] Entity filter (only returns accounts when entity=accounts)
- [x] No results (returns total_count=0)
- [x] Missing query (400)
- [x] No auth (401)

### Not yet implemented
- [ ] Full-text search with Postgres tsvector (deferred — LIKE works for current scale)
- [ ] Advanced filtering UI (deferred)
- [ ] Saved searches / views (deferred)

## Design Decisions

- **LIKE-based search**: Uses SQL LIKE for simplicity and cross-database compatibility (works with both Postgres and SQLite for tests). Can be upgraded to tsvector for performance at scale.
- **Cross-entity**: Single endpoint searches all entities simultaneously, returning grouped results.
- **Entity filter**: Optional `entity` query param to narrow search to one entity type.
- **Authorization**: Search results respect same access control as list endpoints.

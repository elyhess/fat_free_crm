### Advanced Search & Filtering

**Full-Text Search (tsvector)**
- [x] Go migration: add tsvector columns + GIN indexes on accounts, contacts, leads, opportunities, campaigns
- [x] Trigger to auto-update tsvector on insert/update
- [x] Update search handler to use ts_rank + to_tsquery with ILIKE fallback
- [x] Tests (5 existing search tests pass with tsvector)

**Advanced Filtering on Entity Lists**
- [x] Go: add filter query params to entity list endpoints (filter[field_op]=value)
- [x] Supported operators: eq, cont (ILIKE), gt, lt, blank, present
- [x] Column allowlist per entity to prevent SQL injection
- [x] React: FilterDef system on EntityList component
- [x] Filter bars on Accounts, Leads, Opportunities, Campaigns pages
- [x] Entity type dropdown on search results page
- [x] Tests (5 filter tests)

**Saved Searches**
- [x] Go migration: create saved_searches table (user_id, name, entity, filters JSONB)
- [x] Go: CRUD endpoints (list/create/update/delete) scoped to current user
- [x] Tests (3 saved search tests including user isolation)

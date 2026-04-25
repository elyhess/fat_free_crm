### Entity Relationship Endpoints

**Go Backend**
- [x] Join table models: AccountContact, AccountOpportunity, ContactOpportunity
- [x] Repository methods for relationship queries (with auth scoping + pagination)
- [x] Handler: GET /accounts/{id}/contacts
- [x] Handler: GET /accounts/{id}/opportunities
- [x] Handler: GET /campaigns/{id}/leads
- [x] Handler: GET /campaigns/{id}/opportunities
- [x] Handler: GET /contacts/{id}/opportunities
- [x] Route registration in router.go
- [x] Tests for all 5 endpoints (13 tests covering data, pagination, auth, soft-deletes)

**React Frontend**
- [x] RelatedEntities component (reusable, paginated table with links)
- [x] Related entities section on Account detail page (contacts + opportunities)
- [x] Related entities section on Campaign detail page (leads + opportunities)
- [x] Related entities section on Contact detail page (opportunities)

**Verified**
- [x] Authorization scoping enforced on related entities
- [x] Pagination working
- [x] Soft-deleted join records excluded

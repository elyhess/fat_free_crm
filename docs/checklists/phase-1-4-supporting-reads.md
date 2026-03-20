# Phase 1.4 — Supporting Reads

## Checklist

### Models
- [x] Comment model (polymorphic)
- [x] Address model (polymorphic, with soft delete)
- [x] Tag and Tagging models (acts-as-taggable-on)
- [x] Version model (PaperTrail audit log)

### Repository
- [x] ListComments (by entity type + ID)
- [x] ListAddresses (by entity type + ID, soft-delete filtered)
- [x] ListTagsForEntity (via taggings join)
- [x] ListAllTags (ordered by name)
- [x] ListVersions (by entity type + ID, with limit)
- [x] ListRecentActivity (tracked assets only, with limit)
- [x] ListUsers (non-deleted, ordered by username)

### API Endpoints
- [x] GET /api/v1/{entity}/{id}/comments
- [x] GET /api/v1/{entity}/{id}/addresses
- [x] GET /api/v1/{entity}/{id}/tags
- [x] GET /api/v1/{entity}/{id}/versions
- [x] GET /api/v1/tags (all tags)
- [x] GET /api/v1/activity (recent activity feed)
- [x] GET /api/v1/users (admin only, sensitive fields stripped)

### Tests
- [x] Comments filtered by entity
- [x] Addresses with soft-delete exclusion
- [x] Tags via taggings join
- [x] All tags ordered
- [x] Versions filtered by entity
- [x] Recent activity (tracked assets only)
- [x] Users admin-only (403 for non-admin)
- [x] Invalid entity returns 400

## Design Decisions

- **Polymorphic pattern**: Comments, addresses, tags, versions use `{entity}/{id}/` URL pattern with validation against known entity types.
- **Admin-only users endpoint**: Returns sanitized user objects (no passwords, salts, or sign-in tracking fields).
- **Activity feed filters by tracked assets**: Only Account, Campaign, Contact, Lead, Opportunity versions shown (matching Rails `ASSETS` constant).
- **Configurable limits**: Versions and activity endpoints accept `?limit=N` with sensible defaults and caps.

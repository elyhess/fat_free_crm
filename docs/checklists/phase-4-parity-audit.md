# Phase 4 — Feature Parity Audit

## Core CRUD (All Done)

### Accounts
- [x] List, Detail, Create, Update, Delete
- [x] CSV Export / Import
- [x] Comments, Tags, Addresses (polymorphic)

### Contacts
- [x] List, Detail, Create, Update, Delete
- [x] CSV Export / Import
- [x] vCard Export
- [x] Comments, Tags, Addresses

### Leads
- [x] List, Detail, Create, Update, Delete
- [x] CSV Export / Import
- [x] Reject
- [ ] Convert/Promote (complex — creates Account + Contact + Opportunity)

### Opportunities
- [x] List, Detail, Create, Update, Delete
- [x] CSV Export
- [x] Stage transitions

### Campaigns
- [x] List, Detail, Create, Update, Delete
- [x] CSV Export
- [x] Lead counter management

### Tasks
- [x] List, Detail, Create, Update, Delete
- [x] CSV Export
- [x] Complete / Uncomplete

## Supporting Features

- [x] Authentication (JWT, Devise-compatible password hashing)
- [x] Authorization (Public/Private/Shared, admin bypass, group membership)
- [x] Comments (CRUD, polymorphic, owner/admin auth)
- [x] Tags (add/remove, find-or-create, counter management)
- [x] Addresses (CRUD, polymorphic, soft-delete)
- [x] Audit log / Versions (read — PaperTrail compatible)
- [x] Activity feed (recent versions across tracked entities)
- [x] Cross-entity search (LIKE-based, authorization-scoped)
- [x] Dashboard (task summary by bucket, pipeline by stage)
- [x] Users list (admin-only, sensitive fields stripped)

## Admin Functions

- [x] User management (CRUD, suspend, reactivate, password hashing)
- [x] Group management (CRUD, user membership)
- [x] Field group management (CRUD)
- [ ] Custom field definition management (dynamic column creation)
- [ ] Application settings CRUD

## React Frontend

- [x] Login flow with JWT
- [x] Dashboard (task summary + pipeline)
- [x] Entity list pages (all 6, sortable, paginated)
- [x] Create/Edit forms (modal, all entities)
- [x] Delete confirmations
- [x] Search bar + results page
- [x] Layout with nav, user info, logout

## Remaining Gaps (Not Critical for MVP)

- [ ] Lead conversion flow
- [ ] Email integration (IMAP/SMTP)
- [ ] Background jobs
- [ ] Advanced filtering / saved searches
- [ ] Attachment / file upload system
- [ ] Subscribe/unsubscribe notifications
- [ ] Autocomplete endpoints
- [ ] User self-service (password change, profile edit)
- [x] Entity detail pages (dedicated show pages vs. list-only)
- [ ] Audit trail write (Go writes version records on mutations)
- [ ] Custom field values in entity forms

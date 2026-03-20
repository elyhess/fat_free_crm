# Phase 2.1 — Entity Writes

## Checklist

### Tasks
- [x] POST /api/v1/tasks (create)
- [x] PUT /api/v1/tasks/{id} (update)
- [x] DELETE /api/v1/tasks/{id} (delete)
- [x] PUT /api/v1/tasks/{id}/complete (set completed_at, completed_by)
- [x] PUT /api/v1/tasks/{id}/uncomplete (clear completion)
- [x] Owner/assignee/admin authorization check

### Accounts
- [x] POST /api/v1/accounts (create with access control)
- [x] PUT /api/v1/accounts/{id} (update)
- [x] DELETE /api/v1/accounts/{id} (delete)

### Campaigns
- [x] POST /api/v1/campaigns (create)
- [x] PUT /api/v1/campaigns/{id} (update)
- [x] DELETE /api/v1/campaigns/{id} (delete)

### Leads
- [x] POST /api/v1/leads (create, campaign counter increment)
- [x] PUT /api/v1/leads/{id} (update)
- [x] DELETE /api/v1/leads/{id} (delete, campaign counter decrement)
- [x] PUT /api/v1/leads/{id}/reject (set status to rejected)
- [ ] PUT /api/v1/leads/{id}/convert (promote to Account + Contact + Opportunity — complex, deferred)

### Contacts
- [x] POST /api/v1/contacts (create)
- [x] PUT /api/v1/contacts/{id} (update)
- [x] DELETE /api/v1/contacts/{id} (delete)

### Opportunities
- [x] POST /api/v1/opportunities (create)
- [x] PUT /api/v1/opportunities/{id} (update with stage transitions)
- [x] DELETE /api/v1/opportunities/{id} (delete)

### Tests (20 tests)
- [x] Task CRUD (create, update, delete, complete, uncomplete)
- [x] Task forbidden for non-owner
- [x] Account CRUD
- [x] Lead create, reject
- [x] Campaign create
- [x] Contact create
- [x] Opportunity create, stage update
- [x] All write endpoints require auth (401)
- [x] Validation (missing name returns 422)

## Design Decisions

- **Hard delete**: Matches Rails behavior (destroy, not soft delete via deleted_at).
- **Owner/assignee/admin auth**: Write access requires being the record owner, assignee, or admin.
- **Campaign counter management**: Lead create increments, delete decrements `leads_count`.
- **Lead conversion deferred**: The promote flow creates 3 models (Account, Contact, Opportunity) with permission copying — complex enough to warrant a separate implementation.
- **Default access**: If not specified, defaults to "Public" matching Rails `Setting.default_access`.

# Phase 2.3 — Admin Functions

## Checklist

### User Management
- [x] POST /api/v1/admin/users (create with password hashing, auto-confirm)
- [x] PUT /api/v1/admin/users/{id} (update, including password change)
- [x] DELETE /api/v1/admin/users/{id} (delete, with related-asset protection)
- [x] PUT /api/v1/admin/users/{id}/suspend (set suspended_at)
- [x] PUT /api/v1/admin/users/{id}/reactivate (clear suspended_at)
- [x] Cannot delete or suspend self
- [x] Password complexity validation on create/update

### Group Management
- [x] GET /api/v1/admin/groups (list, ordered by name)
- [x] POST /api/v1/admin/groups (create, with user_ids assignment)
- [x] PUT /api/v1/admin/groups/{id} (update name, replace user memberships)
- [x] DELETE /api/v1/admin/groups/{id} (delete, clean up memberships + permissions)

### Field Group Management
- [x] POST /api/v1/admin/field_groups (create)
- [x] PUT /api/v1/admin/field_groups/{id} (update)
- [x] DELETE /api/v1/admin/field_groups/{id} (delete)

### Not yet implemented
- [ ] Application settings CRUD (complex — many setting types, SMTP, IMAP config)
- [ ] Custom field definition CRUD (complex — dynamic column creation)
- [ ] Field sorting/reordering

### Tests (16 tests)
- [x] CreateUser (with password hashing, auto-confirm)
- [x] CreateUser weak password (422)
- [x] CreateUser non-admin (403)
- [x] UpdateUser (field update)
- [x] SuspendUser
- [x] ReactivateUser
- [x] DeleteUser cannot delete self (422)
- [x] SuspendUser cannot suspend self (422)
- [x] CreateGroup
- [x] ListGroups (ordered by name)
- [x] UpdateGroup
- [x] DeleteGroup
- [x] CreateFieldGroup
- [x] UpdateFieldGroup
- [x] DeleteFieldGroup
- [x] All admin endpoints return 403 for non-admin (12 endpoints)

## Design Decisions

- **Admin check in handler**: Each handler method calls `requireAdmin()` rather than using middleware, keeping it simple and explicit.
- **Password hashing**: Uses same `DigestPassword` with random salt, matching existing Devise-compatible auth.
- **Auto-confirm**: Admin-created users get `confirmed_at` set immediately (no email verification).
- **Self-protection**: Cannot delete or suspend yourself to prevent admin lockout.
- **Related-asset protection**: Cannot delete users who own accounts, campaigns, leads, contacts, opportunities, comments, or tasks.
- **Group membership management**: UpdateGroup replaces all memberships (full replace, not incremental).
- **Settings/custom fields deferred**: Application settings and custom field definition management are complex features deferred to a later iteration.

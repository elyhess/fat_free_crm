# Phase 0.4 — Authorization

## Checklist

- [x] Permission and Group models mapping Rails schema
- [x] Access control service (Public/Private/Shared logic)
- [x] Admin bypass (can manage all)
- [x] Owner/assignee check for Private records
- [x] Shared record check via permissions table (user + group)
- [x] Query scope builder for filtered entity lists
- [ ] Authorization middleware (deferred — will wire into router when entity endpoints are added in Phase 1)
- [x] Tests for all of the above

## Design Decisions

- **No Casbin**: The access rules are specific enough that a custom implementation is simpler and more testable than a generic policy engine.
- **Three access levels**: Public (everyone), Private (owner/assignee/admin), Shared (owner/assignee/permitted users+groups/admin).
- **SQL scope builder**: For list endpoints, build WHERE clauses that combine all access conditions in a single query (no N+1).
- **Admin bypass**: Admin users skip all access checks.

## Access Rules (from Rails CanCanCan)

| Access | Who can manage |
|--------|---------------|
| Public | Any authenticated user |
| Private | Creator (`user_id`), assignee (`assigned_to`), admin |
| Shared | Creator, assignee, users/groups in `permissions` table, admin |

Tasks use a different model: accessible by creator, assignee, or completer only.

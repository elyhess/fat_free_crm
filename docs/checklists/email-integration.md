### Email Integration

**Phase 1: Email Model + CRUD**
- [x] Email model (polymorphic mediator — accounts, contacts, leads, opportunities, campaigns)
- [x] GET /{entity}/{id}/emails — list emails for an entity
- [x] DELETE /emails/{id} — delete an email
- [x] React: email timeline on entity detail pages
- [x] Tests (4 handler tests, 7 service tests for entity matching + reply extraction)

**Phase 2: Outgoing Notifications**
- [x] Assignment notification (when user assigned to entity)
- [x] Comment subscription notification (email when comment on subscribed entity)
- [x] Notification emails include entity reference [entity:id] for reply support
- [x] Dropbox notification (email attached to entity)

**Phase 3: IMAP Dropbox + Background Jobs**
- [x] Background job infrastructure (goroutine-based scheduler)
- [x] IMAP polling job — connect, fetch unread, process
- [x] Email dropbox processing: match emails to entities by keyword/recipient/address
- [x] Comment reply processing: parse [entity:id] from subject, create comment
- [ ] IMAP folder management (archive processed, move invalid)
- [ ] Admin settings for dropbox/comment reply IMAP config

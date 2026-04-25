# Phase 2.2 — Supporting Writes

## Checklist

### Comments
- [x] POST /api/v1/{entity}/{id}/comments (create, polymorphic)
- [x] DELETE /api/v1/comments/{id} (delete, owner/admin only)
- [x] Validation (empty comment returns 422)
- [x] Authorization (non-owner, non-admin returns 403)

### Tags
- [x] POST /api/v1/{entity}/{id}/tags (add, find-or-create tag + tagging)
- [x] DELETE /api/v1/{entity}/{id}/tags/{tag_id} (remove tagging, decrement counter)
- [x] Idempotent duplicate tag add (returns 200)
- [x] taggings_count counter management

### Addresses
- [x] POST /api/v1/{entity}/{id}/addresses (create, polymorphic)
- [x] DELETE /api/v1/addresses/{id} (soft-delete filtered lookup + delete)

### Tests (11 new tests)
- [x] CreateComment (polymorphic type set correctly)
- [x] CreateComment empty body (422)
- [x] DeleteComment
- [x] DeleteComment forbidden for non-owner (403)
- [x] AddTag (find-or-create)
- [x] AddTag duplicate (idempotent 200)
- [x] RemoveTag
- [x] CreateAddress (polymorphic type set correctly)
- [x] DeleteAddress
- [x] CreateComment invalid entity (400)
- [x] All supporting write endpoints require auth (401)

### Not yet implemented
- [ ] Audit trail generation (Go writes version records on every mutation)
- [ ] Custom field values (create, update per entity)

## Design Decisions

- **Comment ownership**: Only the comment author or an admin can delete a comment.
- **Tag find-or-create**: If a tag with the given name doesn't exist, it's created. If the tagging already exists, returns 200 (idempotent).
- **taggings_count management**: Incremented on tag add, decremented (with floor of 0) on tag remove.
- **Address soft-delete**: DeleteAddress looks up with `deleted_at IS NULL` filter, matching Rails behavior.
- **Polymorphic routing**: Reuses `validPolymorphicTypes` map from supporting reads to convert URL slugs to Rails class names.

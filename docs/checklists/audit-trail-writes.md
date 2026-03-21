# Audit Trail Writes Checklist

Go writes PaperTrail-compatible version records on every mutation to tracked entities.

## Implementation

- [x] VersionRecorder service (create, update, destroy events)
- [x] JSON serialization for object state and changes
- [x] Whodunnit stored as string user ID (PaperTrail compatible)
- [x] Old object state captured before updates

## Entity Integration

- [x] Account — create, update, delete
- [x] Campaign — create, update, delete
- [x] Contact — create, update, delete
- [x] Lead — create, update, delete, reject
- [x] Opportunity — create, update, delete
- [x] Task — intentionally NOT tracked (matches Rails)

## Tests

- [x] Unit tests: VersionRecorder (5 tests — create, update, destroy, timestamp, multiple)
- [x] Integration tests: handler versioning (8 tests — all tracked entities + task exclusion)
- [x] All existing tests still pass

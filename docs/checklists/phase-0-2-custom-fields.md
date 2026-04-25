# Phase 0.2 — Custom Fields System

## Checklist

- [x] Go models for `field_groups` and `fields` tables (read existing Rails schema)
- [x] Repository layer to query field groups and fields by entity type
- [x] Service layer with field type registry matching Rails `BASE_FIELD_TYPES`
- [x] Validation logic (required, minlength, maxlength)
- [ ] Paired date range validation (deferred — no paired fields exist in DB yet)
- [ ] Custom field value reading from entity tables (dynamic `cf_*` columns) — deferred to entity read phase
- [x] API endpoints: list field groups/fields by entity type
- [x] Tests for all of the above (30 tests passing)
- [x] Document the approach

## Design Decisions

- **Read-first approach**: The Go service reads the existing `field_groups` and `fields` tables that Rails manages. No schema changes needed.
- **No dynamic ALTER TABLE in Go (yet)**: Field creation/modification stays in Rails admin for now. Go reads and validates against the existing field definitions.
- **Field values**: Read `cf_*` columns from entity tables using dynamic SQL (GORM's raw query or map scanning), since Go structs can't have fields added at runtime.
- **Paired fields**: Linked via `pair_id` foreign key. Validate start <= end at the service layer.

## Architecture

```
handler/fields.go          → HTTP handlers (GET /api/v1/field_groups?entity=Account)
service/custom_fields.go   → Field type registry, validation logic
repository/field_groups.go → DB queries for field_groups + fields
model/field.go             → GORM models matching Rails schema
```

## Rails Schema Reference

- `field_groups`: id, name, label, position, hint, tag_id, klass_name
- `fields`: id, type (STI), field_group_id, position, name, label, hint, placeholder, as, collection, disabled, required, maxlength, minlength, pair_id, settings, pattern
- Field types: string, text, email, url, tel, select, radio_buttons, check_boxes, boolean, date, datetime, decimal, integer, float, date_pair, datetime_pair
- `cf_*` columns are added directly to entity tables (accounts, contacts, leads, opportunities, campaigns)

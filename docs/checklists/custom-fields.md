### Custom Field System (End-to-End)

**Go Backend — Reading Custom Fields**
- [x] Read cf_* columns dynamically from entity tables (raw SQL SELECT)
- [x] GET /{entity}/{id}/custom_fields — return cf_* values as JSON
- [x] ReadCustomFieldValuesForList — batch read for entity lists (service layer)

**Go Backend — Writing Custom Fields**
- [x] PUT /{entity}/{id}/custom_fields — update cf_* column values
- [x] Validate custom field values against field definitions (required, minlength, maxlength)
- [x] Write cf_* column values via raw SQL UPDATE
- [x] Reject unknown column names, SQL injection protection (regex validation)

**Go Backend — Admin CRUD for Field Definitions**
- [x] POST /admin/fields — create field definition + ALTER TABLE ADD COLUMN
- [x] PUT /admin/fields/{id} — update field definition (label, hint, required, etc.)
- [x] DELETE /admin/fields/{id} — delete field definition + ALTER TABLE DROP COLUMN
- [x] POST /admin/fields/sort — reorder fields within a group
- [x] Generate cf_* column names from labels (collision-safe: cf_revenue, cf_revenue_2)
- [x] Block modifications to CoreField type fields
- [x] Safe type transition checks (string→text OK, string→integer blocked)

**Go Backend — Tests**
- [x] TestCreateField — create + verify DB column added
- [x] TestCreateField_NonAdmin — 403
- [x] TestCreateField_InvalidType — 422
- [x] TestCreateField_CollisionAvoidance — cf_revenue → cf_revenue_2
- [x] TestUpdateField — update label + required
- [x] TestUpdateField_CoreFieldBlocked — 422
- [x] TestDeleteField — delete + verify column dropped
- [x] TestDeleteField_CoreFieldBlocked — 422
- [x] TestSortFields — reorder positions
- [x] TestGetEntityCustomFields — read cf_* values
- [x] TestUpdateEntityCustomFields — write cf_* values
- [x] TestUpdateEntityCustomFields_UnknownField — 422
- [x] TestUpdateEntityCustomFields_Validation — required field validation
- [x] TestGenerateColumnName — 6 cases
- [x] TestCheckTypeTransition — 7 type pairs
- [x] Service: TestGenerateColumnName, TestIsValidColumnName — column name safety

**React Frontend**
- [x] CustomFieldsDisplay component — renders cf_* values in entity detail pages
- [x] CustomFieldsForm component — dynamic form inputs for all 14 field types
- [x] entityType prop added to all 6 entity detail pages
- [x] AdminFieldsPage — tabbed by entity type, field table, create/edit/delete modals
- [x] Route at /admin/fields, nav link for admin users

**Verified**
- [x] Column types match Rails BASE_FIELD_TYPES (16 types mapped)
- [x] cf_* column naming matches Rails convention
- [x] TypeScript compiles clean
- [x] All new Go tests pass (pre-existing TestTaskSummary_WithTasks failure unrelated)

# Phase 3.2 — Import / Export

## Checklist

### CSV Export
- [x] GET /api/v1/accounts/export (CSV with authorization scope)
- [x] GET /api/v1/contacts/export
- [x] GET /api/v1/leads/export
- [x] GET /api/v1/opportunities/export
- [x] GET /api/v1/campaigns/export
- [x] GET /api/v1/tasks/export (owner/assignee filtered)
- [x] Proper Content-Type and Content-Disposition headers

### CSV Import
- [x] POST /api/v1/accounts/import (CSV body or multipart file upload)
- [x] POST /api/v1/contacts/import
- [x] POST /api/v1/leads/import
- [x] Case-insensitive header matching with aliases (e.g., "First Name", "first_name", "firstname")
- [x] Row-level error reporting (skips bad rows, reports errors)
- [x] GET /api/v1/{entity}/import/template (empty CSV with headers)

### vCard
- [x] GET /api/v1/contacts/export/vcard (vCard 3.0 format)

### Tests (8 tests)
- [x] ExportAccounts (CSV content, headers, data)
- [x] ExportContacts (CSV content)
- [x] ExportAccounts no auth (401)
- [x] ImportAccounts (2 rows imported)
- [x] ImportAccounts with missing name (partial import + error)
- [x] ImportContacts
- [x] ImportTemplate (headers-only CSV)
- [x] VCardExport (vCard format)

## Design Decisions

- **Authorization-scoped exports**: Exports respect the same access control as list endpoints (admin sees all, non-admin sees only accessible records).
- **Flexible CSV parsing**: Headers are case-insensitive and support multiple aliases ("First Name" / "first_name" / "firstname").
- **Partial import**: Bad rows are skipped with errors, good rows are still imported. Response includes count of imported + error messages.
- **No external dependencies**: Uses Go stdlib `encoding/csv` for CSV and manual vCard string building (no go-vcard dependency needed for simple export).
- **Default access**: Imported records default to "Public" access, owned by the importing user.

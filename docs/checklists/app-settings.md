### Application Settings

**Go Backend**
- [x] GET /admin/settings — return all settings as JSON (admin only)
- [x] PUT /admin/settings — bulk update settings (admin only)
- [x] Setting model (name, value with YAML deserialization)
- [x] Type-aware serialization (strings, booleans, arrays, hashes)
- [x] YAML symbol key stripping for clean JSON output
- [x] Tests: 10 tests (GET, PUT, non-admin forbidden, no auth, empty, array update, hash update, YAML deserialization subtests)

**React Frontend**
- [x] Admin settings page with grouped form sections (General, Options, Dropdowns, SMTP, Email Dropbox)
- [x] Boolean toggles, text inputs, textarea for arrays, nested hash forms
- [x] Route at /admin/settings and nav link for admin users

**Verified**
- [x] YAML value column compatibility with Rails serialize :value
- [x] Ruby symbol colon-prefix keys stripped from hash keys in API output
- [x] Round-trip: read YAML from DB → JSON API → update back to YAML

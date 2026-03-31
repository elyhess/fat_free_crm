### User Avatar

**Go Backend**
- [x] Avatar model (maps to existing `avatars` table)
- [x] POST /profile/avatar — multipart file upload (PNG, JPEG, GIF), max 5MB
- [x] DELETE /profile/avatar — remove avatar
- [x] GET /avatars/{user_id} — serve avatar image, Gravatar fallback redirect
- [x] Store files on disk in uploads/avatars/ directory
- [x] avatar_url in profile response
- [x] Tests (11 tests: upload, replace, delete, serve, Gravatar fallback, no auth, invalid type, missing file, too large, no user, profile includes URL)

**React Frontend**
- [x] Avatar display in navbar (28px circle next to username)
- [x] Avatar section on profile page (80px circle with upload/remove buttons)
- [x] File upload via hidden input + FormData POST
- [x] Gravatar fallback when no custom avatar

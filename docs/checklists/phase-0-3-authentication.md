# Phase 0.3 — Authentication

## Checklist

- [x] User model mapping Rails users table
- [x] Password verification (authlogic_sha512: SHA-512 x20 with salt) — verified against real Rails hash
- [x] JWT token generation and validation
- [x] Login endpoint (POST /api/v1/auth/login)
- [x] Auth middleware for protected routes
- [x] Password complexity validation (14+ chars, digit, lower, upper, symbol)
- [x] User status checks (confirmed, not suspended)
- [x] Tests for all of the above (55 total tests passing)

## Design Decisions

- **Password hashing**: authlogic_sha512 — `SHA512(password + salt)` iterated 20 times.
- **JWT**: Stateless tokens with configurable expiry. No session table needed.
- **User lookup**: Case-insensitive match on both `username` and `email` fields (matches Devise behavior).
- **Status checks**: User must have `confirmed_at IS NOT NULL` and `suspended_at IS NULL` (or `sign_in_count > 0`).

## Rails Password Algorithm

```
digest = password + salt
20.times { digest = SHA512.hexdigest(digest) }
```

Constant-time comparison for verification.

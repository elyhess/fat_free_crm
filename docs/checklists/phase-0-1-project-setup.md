# Phase 0.1 — Project Setup

## Checklist

- [x] Initialize Go module and repo structure
- [x] Choose and configure router (chi)
- [x] Set up database connection (pgx + GORM)
- [x] Configuration management (env vars, config files)
- [x] Docker Compose for local dev (Go service + Postgres)
- [x] Logging and error handling patterns
- [x] Health check endpoint with tests
- [x] CI pipeline (golangci-lint, tests)

## Decisions

- **Router:** chi — stdlib-compatible, idiomatic Go
- **ORM:** GORM with pgx driver — closest to ActiveRecord (associations, callbacks, migrations)
- **Config:** Environment variables with a `.env` file for local dev
- **Logging:** Go stdlib `slog` (structured logging, zero dependencies)

## Notes

- The Go service connects to the same Postgres database as Rails (`fat_free_crm_development`).
- Go service runs on port 8080 (Rails stays on 3000).

### Database Migration Ownership

- [x] Add goose dependency to Go project
- [x] Generate baseline migration from current Rails schema (pg_dump --schema-only)
- [x] Add migrate command (cmd/migrate) with up, down, status, create, mark-applied
- [x] Mark baseline as applied on existing dev database
- [x] Add Makefile targets (migrate, migrate-status, migrate-down, migrate-create, migrate-mark-baseline)
- [x] Document migration workflow in docs/local-development.md
- [x] Test: fresh database from baseline produces working schema

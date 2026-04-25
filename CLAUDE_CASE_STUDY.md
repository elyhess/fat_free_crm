# Case Study: Migrating Fat Free CRM with Claude Opus

**Rails 7 → Go + React, end-to-end, in roughly two weeks of working sessions.**

---

## The Headline

| | |
|---|---|
| **Source app** | Fat Free CRM — a 16-year-old Ruby on Rails CRM (~50k LoC of legacy Rails 7.1, Haml views, CoffeeScript, Devise, CanCanCan, Ransack, PaperTrail) |
| **Target stack** | Go backend (chi + GORM + Postgres) + React SPA (Vite + TypeScript + Tailwind), shipped as a single `go:embed` binary |
| **Calendar time** | First Go commit **2026-03-20**, Rails decommissioned **2026-04-03** (~15 days, working sessions only) |
| **Coverage** | **208 of 215 application routes** ported or replaced (96.7%) — verified by a 274-route audit and a 93-endpoint smoke test |
| **What Claude wrote** | ~19,800 LoC of Go (production) + ~10,300 LoC of Go tests + ~5,500 LoC of React/TypeScript + ~3,500 LoC of Playwright E2E tests |
| **Test ratio** | ~52% test code to production code in the Go backend, plus a full E2E suite |

This is a real working application, not a toy. Auth, authorization rules, audit trail, custom fields, IMAP, full-text search, file uploads, CSV import/export — all migrated and verified against the original Rails behavior.

---

## How the Work Happened

The migration ran on a disciplined loop, codified in `CLAUDE.md` and `docs/go_migration_plan.md`:

1. Pick the next unchecked item from the phased plan.
2. Read the existing Rails code (model, controller, views, specs).
3. Write Go tests first, then implement the handler/service/repository.
4. Build the matching React component.
5. Verify parity against the running Rails app.
6. Commit the code **and** the updated checklist together on a feature branch.

Each of the ~30 feature branches in this repo (`go-react-refactor/<feature>`) is a single Claude session that produced one shippable slice. The checklists in `docs/checklists/` are the artifacts of those sessions.

**What this looked like in practice — the first day (2026-03-20):**

| Time | Commit |
|---|---|
| 14:11 | Go backend project scaffold (Phase 0.1) |
| 14:19 | Custom fields system (Phase 0.2) |
| 14:25 | Authentication — Devise-compatible (Phase 0.3) |
| 14:31 | Authorization service (Phase 0.4) |
| 14:36 | React frontend scaffold (Phase 1.1) |
| 14:42 | Read-only API for all 6 CRM entities (Phase 1.3) |
| 15:04 | Dashboard endpoints (Phase 1.2) |
| 15:08 | React entity list views + live dashboard |
| 15:16 | Comments, addresses, tags, versions, users (Phase 1.4) |
| 15:27 | Write endpoints for all CRM entities (Phase 2.1) |
| 15:33 | Comments, tags, addresses (writes) |
| 15:41 | Admin functions: users, groups, field groups |
| 15:50 | React create/edit/delete forms for every entity |

**That is Phases 0, 1, and 2 — the full read and write surface of the CRM, plus a working React UI — in a single afternoon.**

---

## Four Decisions Worth Showing Your Boss

These are non-obvious calls Claude proposed and executed. Each one would normally be a multi-day spike for a senior engineer.

### 1. Custom fields without a schema rewrite

**Problem:** Rails' custom-field system uses two tables (`fields`, `field_groups`) and dynamically `ALTER TABLE`s the entity tables to add `cf_*` columns at runtime. The "obvious" Go rewrite is a JSONB migration, which means a one-way data move and downtime.

**Decision (Claude):** *Reuse the existing tables and dynamic columns.* The Go backend reads `cf_*` columns into a `map[string]any` and exposes admin endpoints that issue safe `ALTER TABLE` statements with collision-safe naming and validated type transitions.

**Why it matters:** Zero data migration. Rails and Go could read/write the same custom fields in parallel during the cutover. Marked in the plan as the "architectural linchpin — validate the Go approach before building on top of it."

### 2. Devise-compatible password hashing (cryptographic interop)

**Problem:** Existing user passwords are hashed by `devise-encryptable` using `authlogic_sha512` — Devise salt + 20 rounds of SHA-512 with a specific concatenation order. Asking users to reset their passwords on cutover was a non-starter.

**Decision (Claude):** Reproduce the exact `authlogic_sha512` algorithm in Go and verify against the live Postgres rows. Login endpoint accepts either username or email, runs the same complexity checks Devise enforced, and respects `confirmed_at` / `suspended_at`.

**Result:** Every existing user logs in with their existing password on day one. No reset flow, no support tickets.

### 3. PaperTrail-compatible audit trail (write side)

**Problem:** PaperTrail writes serialized YAML into a `versions` table. The Rails app reads that table to render entity history pages. If Go writes mutations without producing matching version rows, the audit history breaks the moment Go takes over.

**Decision (Claude):** Generate version records on every Go mutation in the exact format PaperTrail expects (YAML-serialized object snapshots, polymorphic `item_type` / `item_id`, correct event names). Now both old (Rails-written) and new (Go-written) versions render identically in the UI.

### 4. Single-binary deployment via `go:embed`

**Problem:** Rails-style deployment means Puma + asset pipeline + Nginx. Standard "Go API + separate React static host" still means two deployable artifacts.

**Decision (Claude):** Build the React production bundle, copy it into `go-backend/internal/frontend/dist/`, and `go:embed` it into the Go binary. The chi router serves API routes and falls back to the SPA's `index.html` for client-side routes. `make build-frontend` automates the whole chain.

**Result:** The entire app — backend, frontend, all assets — ships as a single statically-linked binary. No runtime Node, no separate static host.

---

## Process Discipline Claude Maintained

This is the part senior engineers care about. Claude wasn't just typing code:

- **Test-first, always.** The Go backend has a ~52% test-to-production ratio — written in the same session as the feature, not bolted on later. Plus a 93-endpoint smoke test suite (`go-backend/scripts/smoke-test.sh`) and a Playwright E2E suite (~3,500 LoC) that drives the running app.
- **One concern per branch.** Every feature lives on its own `go-react-refactor/<feature>` branch with a matching checklist file in `docs/checklists/`. The git history is a clean, reviewable record of the migration.
- **A formal parity audit before decommission.** `docs/route-audit.md` walks all 274 Rails routes and classifies each as Ported, Replaced by React, N/A (Rails framework internals), or Not Ported (with written justification). Nothing was forgotten.
- **Conservative dependency choices.** Claude proposed and executed *not* using Casbin (custom authorization service was simpler), *not* using Redis (goroutine scheduler was enough), and *not* upgrading to `tsvector` until LIKE search hit its limits. Each "no" kept the stack smaller.
- **Knew when to defer.** The plan has explicit `[ ]` items marked as deferred with reasons (e.g. "Lists CRUD — minor feature superseded by saved searches"). Scope was managed, not abandoned.

---

## The Pitch

A two-engineer team spending six months on this migration would be a normal estimate — it touches every layer of the stack, and the cryptographic + audit-trail interop alone are easy to get wrong. Claude Opus, paired with one operator, did it in roughly **fifteen working days of calendar time**, with tests, documentation, parity audit, and a clean commit history.

The bottleneck was no longer "can we build it" — it was "what do we want next." That is the shift worth showing.

---

## Where to Look in the Repo

- `docs/go_migration_plan.md` — the phased plan we executed against
- `docs/route-audit.md` — the 274-route Rails-vs-Go parity audit
- `docs/remaining-work.md` — final status snapshot
- `docs/checklists/` — one file per feature branch, the per-session record
- `go-backend/scripts/smoke-test.sh` — the 93-endpoint verification script
- `e2e/tests/` — Playwright suite covering auth, CRUD, admin, dashboard, search
- `git log --oneline go-react-refactor` — the migration in commit form

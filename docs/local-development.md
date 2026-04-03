# Local Development

How to run the Go backend and React frontend locally.

## Prerequisites

- **Go** 1.26+ (see `go-backend/go.mod`)
- **Node.js** 20+ and npm
- **PostgreSQL** 16+ (running locally or via Docker)

## Database

### Option A: Local PostgreSQL

```bash
createdb fat_free_crm_development
```

### Option B: Docker

```bash
cd go-backend
docker compose up -d db
```

This starts PostgreSQL on port 5432 with trust auth and creates the `fat_free_crm_development` database.

### Schema Migrations

The Go backend manages schema via [goose](https://github.com/pressly/goose). Migration files live in `go-backend/db/migrations/`.

**Fresh database (no existing Rails schema):**

```bash
cd go-backend
make migrate          # applies all migrations including baseline
```

**Existing database (already has Rails schema):**

```bash
cd go-backend
make migrate-mark-baseline   # marks baseline as applied without running it
make migrate                 # applies any subsequent migrations
```

**Other migration commands:**

```bash
make migrate-status                  # show which migrations are applied
make migrate-down                    # roll back the last migration
make migrate-create NAME=add_foo     # create a new migration file
```

### Seed Data

Seed data is currently managed by the Rails app. See the main `CLAUDE.md` for Rails setup instructions to run `db:seed` and `ffcrm:demo:load`.

## Go Backend

The Go API server runs on port **8080** by default.

### Configuration

Copy the env example and adjust if needed:

```bash
cd go-backend
cp .env.example .env
```

**Important:** The Go app defaults to `fat_free_crm_development`, but Rails may use a different name (e.g. `fat_free_crm_<username>_development`). Check `config/database.yml` and set `DB_DATABASE` accordingly in your `.env`.

Environment variables (all have sensible defaults for local dev):

| Variable       | Default                        | Description                          |
|----------------|--------------------------------|--------------------------------------|
| `PORT`         | `8080`                         | HTTP listen port                     |
| `DATABASE_URL` | *(built from DB_* vars below)* | Full Postgres connection string      |
| `DB_HOST`      | `localhost`                    | Postgres host                        |
| `DB_PORT`      | `5432`                         | Postgres port                        |
| `DB_USERNAME`  | *(empty — uses OS user)*       | Postgres user                        |
| `DB_PASSWORD`  | *(empty)*                      | Postgres password                    |
| `DB_DATABASE`  | `fat_free_crm_development`     | Database name                        |
| `DB_SSLMODE`   | `disable`                      | Postgres SSL mode                    |
| `JWT_SECRET`   | `change-me-in-production`      | HMAC secret for JWT signing          |
| `LOG_LEVEL`    | `info`                         | Log level: debug, info, warn, error  |

### Run

```bash
cd go-backend
go run ./cmd/server
```

The server starts at `http://localhost:8080`. Verify with:

```bash
curl http://localhost:8080/health
```

### Run Tests

```bash
cd go-backend
make test             # runs all tests (uses -p 1 for DB safety)
make test-verbose     # verbose output
```

Tests use the PostgreSQL test database (`fat_free_crm_elyhess_test`). The `-p 1` flag serializes packages that share the same database. See `CLAUDE.md` for more details.

## React Frontend

The Vite dev server runs on port **3000** and proxies `/api/*` requests to the Go backend at `localhost:8080`.

### Install Dependencies

```bash
cd react-frontend
npm install
```

### Run

```bash
cd react-frontend
npm run dev
```

Open `http://localhost:3000` in your browser.

### Login

Use the demo credentials:

| Field    | Value              |
|----------|--------------------|
| Username | `admin`            |
| Password | `Dem0P@ssword!!`   |

### Build for Production

```bash
npm run build     # outputs to dist/
npm run preview   # preview the production build
```

### Lint & Type Check

```bash
npm run lint          # ESLint
npx tsc --noEmit      # TypeScript check
```

## Single Binary (Production Build)

Build the React frontend and embed it in the Go binary:

```bash
cd go-backend
make build          # runs build-frontend, then builds Go binaries
```

This:
1. Builds the React app (`npm run build` in react-frontend/)
2. Copies `dist/` into `go-backend/internal/frontend/dist/`
3. Compiles the Go binary with the frontend embedded via `go:embed`

Run the single binary:

```bash
./bin/server
```

The server serves both the API (`/api/v1/*`, `/health`) and the React SPA (all other routes) on port 8080.

## Running for Development

Open three terminals:

```bash
# Terminal 1 — Database (skip if Postgres is already running)
cd go-backend && docker compose up -d db

# Terminal 2 — Go backend
cd go-backend && go run ./cmd/server

# Terminal 3 — React frontend
cd react-frontend && npm run dev
```

Then open `http://localhost:3000`.

> **Note:** During development, use the Vite dev server (port 3000) for hot reload. The Go server still serves the embedded frontend on port 8080, but it only updates when you run `make build-frontend`.

## Docker Compose (Full Stack)

To run the Go backend and PostgreSQL together via Docker:

```bash
cd go-backend
docker compose up
```

This starts both the `db` (Postgres on port 5432) and `api` (Go on port 8080) services. The React frontend still needs to be run separately with `npm run dev`.

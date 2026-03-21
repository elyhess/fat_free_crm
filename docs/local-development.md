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

### Seed Data

The database schema and seed data are currently managed by the Rails app. See the main `CLAUDE.md` for Rails setup instructions to run migrations and seed the database.

## Go Backend

The Go API server runs on port **8080** by default.

### Configuration

Copy the env example and adjust if needed:

```bash
cd go-backend
cp .env.example .env
```

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
go test ./...
```

Tests use an in-memory SQLite database — no external dependencies needed.

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

## Running Everything Together

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

## Docker Compose (Full Stack)

To run the Go backend and PostgreSQL together via Docker:

```bash
cd go-backend
docker compose up
```

This starts both the `db` (Postgres on port 5432) and `api` (Go on port 8080) services. The React frontend still needs to be run separately with `npm run dev`.

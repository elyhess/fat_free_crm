# AGENTS.md

Guidance for agents working in this repository.

## Project Snapshot

- Fat Free CRM is a Rails 7.1 application (Ruby 3.4, Bundler 2.4).
- An open source, Ruby on Rails customer relationship management platform.
- The UI is server-rendered and legacy Rails-flavored: Haml views, CoffeeScript, SCSS, `form_for`, `respond_with`, and AJAX templates such as `*.js.haml` and `*.js.erb`.
- Database: **PostgreSQL** (the `pg` gem is the default in the Gemfile).
- Authentication: Devise (~4.6) with `devise-security` and `devise-encryptable`.
- Authorization: CanCanCan.
- Key gems: PaperTrail (auditing), will_paginate, simple_form, acts-as-taggable-on, Ransack (search), Bootstrap 5.2.

## Architecture

### Core Entities (`app/models/entities/`)

The CRM revolves around these models:

- **Account** - companies/organizations
- **Contact** - people associated with accounts
- **Lead** - potential contacts/sales prospects
- **Opportunity** - deals/revenue tracked against accounts
- **Campaign** - marketing campaigns that generate leads

Join models: `AccountContact`, `AccountOpportunity`, `ContactOpportunity`.

### Custom Fields

- `app/models/fields/`, `lib/fat_free_crm/fields.rb`, `lib/fat_free_crm/custom_fields.rb`
- Dynamic field system that lets admins add custom attributes to entities at runtime.

### Key Directories

| Directory | Purpose |
|---|---|
| `app/controllers/admin` | Admin UI controllers |
| `app/views/admin` | Admin configuration screens (Haml) |
| `app/models/entities` | Core CRM entity models |
| `app/models/fields` | Custom/dynamic field system |
| `app/models/users` | Auth, authorization, user lifecycle |
| `app/views` | Haml templates |
| `app/assets/javascripts` | CoffeeScript assets |
| `app/assets/stylesheets` | SCSS stylesheets |
| `lib/tasks/ffcrm` | Project rake tasks (setup, etc.) |
| `spec` | RSpec test suite |
| `db/seeds/` | Seed data files loaded by `db/seeds.rb` |

## Demo Login

- **Username:** `admin`
- **Password:** `Dem0P@ssword!!`
- **Email:** `admin@fatfreecrm.local`

Non-admin login:
- **Username:** `demo`
- **Password:** `Dem0P@ssword!!`
- **Email:** `demo@fatfreecrm.local`

Additional demo users: aaron, ben, cindy, dan, elizabeth, frank, george, heather (all admins, passwords unknown — use the admin account above).

## Local Setup

```bash
# 1. Use correct Ruby (3.4.x from .ruby-version)
# 2. Install gems
bundle install

# 3. Set up database config
cp config/database.postgres.yml config/database.yml

# 4. Create databases and run migrations
bundle exec rake db:create
bundle exec rake db:migrate

# 5. Seed the database (run once — not idempotent)
bundle exec rake db:seed

# 6. Load demo data (100+ accounts, contacts, leads, opportunities, campaigns, tasks)
bundle exec rake ffcrm:demo:load

# 7. Create admin user AFTER demo data (demo fixtures overwrite users)
#    Note: macOS sets USERNAME env var to your OS user — override it explicitly.
bundle exec rails runner '
u = User.find_or_initialize_by(username: "admin")
u.email = "admin@fatfreecrm.local"
u.password = "Dem0P@ssword!!"
u.password_confirmation = "Dem0P@ssword!!"
u.admin = true
u.skip_confirmation!
u.confirm
u.save!
u.update_column(:suspended_at, nil)
'
```

### Database Notes

- PostgreSQL is the primary database. Config template: `config/database.postgres.yml`.
- Default databases: `fat_free_crm_development`, `fat_free_crm_test`, `fat_free_crm_production`.
- Create databases before running migrations — app initialization may touch the DB early.
- Seed data for field groups is not idempotent; repeated runs can create duplicate records.

## Running Tests

### Go Backend

```bash
cd go-backend

# Run all tests (uses -p 1 to serialize packages sharing the PG test database)
make test

# Verbose output
make test-verbose

# Run a single test
go test -p 1 ./internal/handler/... -run TestAutocomplete_Accounts -v
```

**Important:** Always use `make test` (or `-p 1`) when running the full suite. Multiple packages share the same PostgreSQL test database (`fat_free_crm_elyhess_test`) and will corrupt each other's data if run in parallel.

### Rails (Legacy)

```bash
# Prepare the test database
RAILS_ENV=test bundle exec rake db:migrate

# Run the full suite
bundle exec rspec

# Run a specific spec file
bundle exec rspec spec/path/to/file_spec.rb

# Run specs by category
bundle exec rake spec:models
bundle exec rake spec:controllers
bundle exec rake spec:views
bundle exec rake spec:helpers
bundle exec rake spec:routing
bundle exec rake spec:mailers
bundle exec rake spec:lib
bundle exec rake spec:features
```

### Linting

```bash
bundle exec rubocop
```

## Editing Guidance

- Prefer small, local changes over broad rewrites.
- Match the style of the file you are touching. This codebase predates many modern Rails conventions.
- Do not rewrite Haml to ERB, CoffeeScript to another frontend stack, or server-rendered flows to Hotwire/Turbo unless the task explicitly requires it.
- Preserve existing controller and view patterns: `respond_with`, remote forms, partial-heavy rendering, JS response templates.
- When changing custom fields or admin configuration screens, check the model, controller, admin partials, and any seed or rake task code that touches the same concept.
- Only update `db/schema.rb` when the task intentionally changes schema.
- Only update `Gemfile.lock` when the task intentionally changes dependencies.

## Go + React Migration Workflow

We are migrating this Rails app to a Go backend with a React frontend. The following workflow must be followed for all migration work.

### Reference Documents

- `docs/go_migration_plan.md` — phased plan with task checklists and entity migration template
- `docs/go_dependency_mapping.md` — Ruby gem → Go library equivalents

### Session Workflow

1. **Pick the next piece of work** from `docs/go_migration_plan.md`.
2. **Create an entity checklist** by copying the per-entity migration checklist template from the plan into a new file at `docs/checklists/<entity_or_feature>.md`. Check items off as they are completed.
3. **Read the existing Rails code** for that piece — model, controller, views, specs — to understand current behavior.
4. **Consult `docs/go_dependency_mapping.md`** when choosing Go libraries or patterns for the implementation.
5. **Write tests first, then implement.** Every piece of functionality must have tests. Run them and confirm they pass before moving on.
6. **Branch and commit workflow:**
   - All migration work branches off `go-react-refactor`.
   - For each piece, create a descriptively named branch: `git checkout -b go-react-refactor/<feature-name> go-react-refactor`
   - Examples: `go-react-refactor/project-scaffold`, `go-react-refactor/auth`, `go-react-refactor/accounts-read-api`
   - Once tests pass, commit the code **and** the updated checklist doc together.
   - Merge back into `go-react-refactor` when complete.
7. **Update the checklist** — mark completed items in both the feature checklist (`docs/checklists/`) and the main plan (`docs/go_migration_plan.md`).

### Rules

- Never skip tests. If there are no tests, the work is not done.
- Keep commits focused — one concern per branch/commit.
- The checklist file is part of the deliverable and must be committed with the code.
- Consult the dependency mapping before adding any new Go dependency.

## Common Gotchas

- Views are **Haml**, not ERB.
- Asset behavior may live in CoffeeScript or JS response templates rather than modern frontend tooling.
- The app contains old and new patterns side by side; do not assume a single architectural style.
- If you change the custom-field schema or model attributes, verify the seed files still match the live attribute names.

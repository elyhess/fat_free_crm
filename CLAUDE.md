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
```

### Database Notes

- PostgreSQL is the primary database. Config template: `config/database.postgres.yml`.
- Default databases: `fat_free_crm_development`, `fat_free_crm_test`, `fat_free_crm_production`.
- Create databases before running migrations — app initialization may touch the DB early.
- Seed data for field groups is not idempotent; repeated runs can create duplicate records.

## Running Tests

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

## Common Gotchas

- Views are **Haml**, not ERB.
- Asset behavior may live in CoffeeScript or JS response templates rather than modern frontend tooling.
- The app contains old and new patterns side by side; do not assume a single architectural style.
- If you change the custom-field schema or model attributes, verify the seed files still match the live attribute names.

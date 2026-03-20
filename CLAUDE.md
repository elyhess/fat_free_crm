# AGENTS.md

Guidance for agents working in this repository.

## Project Snapshot

- Fat Free CRM is a Rails 7.1 application.
- The repo pins Ruby via `.ruby-version` and RuboCop targets Ruby 3.4.
- The UI is server-rendered and legacy Rails-flavored: Haml views, CoffeeScript, SCSS, `form_for`, `respond_with`, and AJAX templates such as `*.js.haml` and `*.js.erb`.
- Database support exists for PostgreSQL, MySQL, and SQLite, but the default Gemfile path uses PostgreSQL unless `DB=sqlite` or `DB=mysql` is set.
- CI currently runs specs against SQLite, not PostgreSQL.

## Read The Code, Not Just The README

- Some top-level docs are stale. Example: the README still describes older Ruby requirements, while the repo itself pins a newer Ruby in `.ruby-version`.
- When setup or behavior matters, prefer `Gemfile`, `Gemfile.lock`, `.github/workflows/`, `config/database.*.yml`, and the relevant models/controllers over prose docs.

## Important Directories

- `app/controllers/admin`, `app/views/admin`: admin UI and configuration screens.
- `app/models/entities`: core CRM records such as accounts, contacts, leads, opportunities, and tasks.
- `app/models/fields`, `lib/fat_free_crm/fields.rb`, `lib/fat_free_crm/custom_fields.rb`: dynamic/custom field system.
- `app/models/users`: authentication, authorization, and user lifecycle logic.
- `app/views`: Haml templates throughout the app.
- `app/assets/javascripts`, `app/assets/stylesheets`: CoffeeScript and SCSS assets.
- `lib/tasks/ffcrm`: project rake tasks, including setup-related tasks.
- `spec`: RSpec test suite split by Rails layer and feature area.

## Local Setup

1. Use the Ruby version from `.ruby-version`.
2. Use the Bundler version recorded in `Gemfile.lock`.
3. Create `config/database.yml` from the appropriate template:
   - `config/database.postgres.yml` for PostgreSQL
   - `config/database.sqlite.yml` for SQLite
4. Install gems with `bundle install`.
5. Create the databases, then run migrations.

### Database Notes

- For local development, PostgreSQL works well and matches the default Gemfile path.
- For CI parity, SQLite is the most accurate local test target because the GitHub workflow runs specs with `DB=sqlite`.
- On a blank PostgreSQL setup, creating the databases before running migrations is safer than relying on `db:prepare`, because app initialization may touch the database early.

## Seeds

- `db/seeds.rb` loads seed files from `db/seeds/`.
- Treat `bin/rails db:seed` as bootstrap setup, not something to run repeatedly without checking the results.
- Seed data for field groups is not written to be strongly idempotent; repeated runs can create duplicate records.
- If you change the custom-field schema or model attributes, verify the seed files still match the live attribute names.

## Editing Guidance

- Prefer small, local changes over broad rewrites.
- Match the style of the file you are touching. This codebase predates many modern Rails conventions.
- Do not rewrite Haml to ERB, CoffeeScript to another frontend stack, or server-rendered flows to Hotwire/Turbo unless the task explicitly requires it.
- Preserve existing controller and view patterns such as `respond_with`, remote forms, partial-heavy rendering, and JS response templates.
- When changing custom fields or admin configuration screens, check the model, controller, admin partials, and any seed or rake task code that touches the same concept.
- Be cautious with generated files:
  - Only update `db/schema.rb` when the task intentionally changes schema.
  - Only update `Gemfile.lock` when the task intentionally changes dependencies or lockfile platforms.

## Testing And Validation

- Prefer targeted tests first.
- Useful commands:
  - `bundle exec rspec spec/path/to/file_spec.rb`
  - `DB=sqlite RAILS_ENV=test bundle exec rake spec:preparedb`
  - `DB=sqlite bundle exec rake spec:models`
  - `DB=sqlite bundle exec rake spec:controllers`
  - `DB=sqlite bundle exec rake spec:views`
  - `DB=sqlite bundle exec rake spec:helpers`
  - `DB=sqlite bundle exec rake spec:routing`
  - `DB=sqlite bundle exec rake spec:mailers`
  - `DB=sqlite bundle exec rake spec:lib`
  - `DB=sqlite bundle exec rake spec:features`
- Run `bundle exec rubocop` for Ruby changes when practical.
- If you touch authentication, admin flows, or custom fields, favor at least one targeted feature, controller, or model spec in addition to linting.

## Common Gotchas

- Views are Haml, not ERB.
- Asset behavior may live in CoffeeScript or JS response templates rather than modern frontend tooling.
- The app contains old and new patterns side by side; do not assume a single architectural style across the whole repository.
- Setup commands that work in SQLite may behave differently in PostgreSQL, especially around initial database creation and seed/bootstrap flows.

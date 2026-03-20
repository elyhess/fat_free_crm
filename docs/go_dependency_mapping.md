# Rails → Go Dependency Mapping

A reference mapping of Ruby gems used in Fat Free CRM to their Go equivalents.

## Framework & Routing

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `rails` | Full MVC framework | [gin](https://github.com/gin-gonic/gin) or [echo](https://github.com/labstack/echo) or [chi](https://github.com/go-chi/chi) | Go doesn't have a monolithic framework equivalent. chi is the most idiomatic (stdlib-compatible). Gin is the most popular. |
| `responders` | Respond to multiple formats (JSON, HTML, XML) | Built-in with gin/echo | Content negotiation is straightforward in Go routers. |
| `puma` | App server | Built-in `net/http` | Go's stdlib HTTP server is production-grade. No separate app server needed. |
| `bootsnap` | Boot time optimization | N/A | Go compiles to a binary — no boot optimization needed. |

## Database & ORM

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `pg` | PostgreSQL driver | [pgx](https://github.com/jackc/pgx) | The de facto Go Postgres driver. High performance, pure Go. |
| ActiveRecord (via Rails) | ORM | [GORM](https://github.com/go-gorm/gorm) or [sqlc](https://github.com/sqlc-dev/sqlc) or [Bun](https://github.com/uptrace/bun) | GORM is closest to ActiveRecord (associations, callbacks, migrations). sqlc generates type-safe code from SQL. Bun is a lighter ORM. |
| `will_paginate` | Pagination | [go-paginate](https://github.com/raphaelvigee/go-paginate) or manual `LIMIT/OFFSET` | Pagination is simple enough to implement directly. GORM has `.Scopes()` for this. |
| `ransack` | Search/filtering | Custom query builder or [gorm-query](https://github.com/nicholasgasior/gorm-query) | No direct equivalent. You'd build a filter struct → SQL translation layer. |
| `acts_as_list` | Ordered lists (position column) | Manual `position` column management | Simple enough to handle with `UPDATE ... SET position = ?`. |
| `database_cleaner` | Test DB cleanup | [testfixtures](https://github.com/go-testfixtures/testfixtures) or transaction rollback in tests | Wrap each test in a transaction and rollback. |

## Authentication & Authorization

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `devise` | Authentication (login, registration, password reset, email confirmation) | [authboss](https://github.com/volatiletech/authboss) or custom with [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) + [JWT](https://github.com/golang-jwt/jwt) | No single Go library matches Devise's breadth. Most Go projects build auth from smaller pieces. |
| `devise-encryptable` | Legacy password encryption | [crypto/sha512](https://pkg.go.dev/crypto/sha512) | For migrating existing password hashes. New passwords should use bcrypt. |
| `devise-security` | Password complexity, expiry | Custom validation | Write password policy as middleware/validator. |
| `cancancan` | Role-based authorization | [casbin](https://github.com/casbin/casbin) or [oso](https://github.com/osohq/go-oso) | Casbin supports RBAC, ABAC, and ACL models. Closest to CanCanCan's flexibility. |

## Auditing & Versioning

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `paper_trail` | Model versioning / audit log | Custom audit table + GORM callbacks or [gorm-audit](https://github.com/ggwhite/go-gorm-audit) | Store diffs in a `versions` table via GORM hooks (`BeforeUpdate`, `AfterCreate`). |
| `rails-observers` | Model lifecycle callbacks | GORM hooks | GORM provides `BeforeSave`, `AfterCreate`, `BeforeDelete`, etc. |

## Tagging, Comments & Associations

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `acts-as-taggable-on` | Polymorphic tagging | Custom `tags` + `taggings` tables | Build a `Tagging` join model with `taggable_type` / `taggable_id`. GORM supports polymorphic associations. |
| `acts_as_commentable` | Polymorphic comments | Custom `comments` table | Same pattern — polymorphic `commentable_type` / `commentable_id`. |

## Email & Communication

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| ActionMailer (via Rails) | Send emails | [gomail](https://github.com/wneessen/go-mail) or stdlib `net/smtp` | go-mail is the most full-featured. |
| `premailer` | Inline CSS in emails | [premailer](https://github.com/vanng822/go-premailer) | Direct Go port exists. |
| `email_reply_parser_ffcrm` | Parse reply content from emails | [emailreplyparser](https://github.com/xeoncross/go-emailreplyparser) | Community port of the same concept. |
| Net::IMAP (stdlib) | IMAP email fetching | [go-imap](https://github.com/emersion/go-imap) | Well-maintained IMAP client. |

## Data Handling

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `nokogiri` | HTML/XML parsing | [goquery](https://github.com/PuerkitoBio/goquery) (HTML) or `encoding/xml` (XML) | goquery provides jQuery-like HTML traversal. |
| `vcardigan` | vCard parsing/generation | [go-vcard](https://github.com/emersion/go-vcard) | vCard encoding/decoding. |
| `csv` | CSV processing | `encoding/csv` (stdlib) | Go's stdlib CSV is solid. |
| `addressable` | URI parsing | `net/url` (stdlib) | Go's stdlib handles URI parsing well. |
| `country_select` | Country list/selection | [countries](https://github.com/biter777/countries) | Country data and ISO codes. |
| `rails_autolink` | Auto-link URLs in text | [bluemonday](https://github.com/microcosm-cc/bluemonday) + regex | bluemonday for HTML sanitization, regex or [xurls](https://github.com/mvdan/xurls) for URL detection. |

## Image Processing

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `mini_magick` / `image_processing` | Image resize/transform | [imaging](https://github.com/disintegration/imaging) or [bimg](https://github.com/h2non/bimg) | imaging is pure Go. bimg wraps libvips for higher performance. |

## Internationalization

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `rails-i18n` / `devise-i18n` | Translation/localization | [go-i18n](https://github.com/nicksnyder/go-i18n) or [gotext](https://pkg.go.dev/golang.org/x/text) | go-i18n supports YAML/JSON translation files similar to Rails. |

## Fake Data & Testing

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `ffaker` | Fake data generation | [gofakeit](https://github.com/brianvoe/gofakeit) | Names, addresses, companies, emails, etc. |
| `factory_bot_rails` | Test fixtures/factories | [factory-go](https://github.com/bluele/factory-go) or struct literals | Most Go projects use plain struct builders rather than a factory DSL. |
| `rspec-rails` | Test framework | `testing` (stdlib) + [testify](https://github.com/stretchr/testify) | testify adds assertions and mocking on top of the stdlib. |
| `capybara` / `selenium-webdriver` | Browser integration tests | [chromedp](https://github.com/chromedp/chromedp) or [rod](https://github.com/nicholasgasior/rod) | For E2E tests against the React frontend. |
| `timecop` | Time mocking in tests | [clock](https://github.com/benbjohnson/clock) | Inject a clock interface instead of calling `time.Now()` directly. |
| `rubocop` | Linter | `go vet` + [golangci-lint](https://github.com/golangci/golangci-lint) | golangci-lint bundles 50+ linters. `gofmt` handles formatting. |

## Deployment

| Ruby Gem | Purpose | Go Equivalent | Notes |
|---|---|---|---|
| `capistrano` | Deployment automation | N/A — deploy the binary | Go compiles to a single binary. Use Docker, systemd, or k8s. |
| `rails_12factor` | Heroku compatibility | N/A | Go apps are already 12-factor friendly by default. |

## No Go Equivalent Needed

| Ruby Gem | Why Not Needed |
|---|---|
| `sprockets-rails`, `sassc-rails`, `coffee-rails`, `uglifier`, `execjs`, `mini_racer` | Asset pipeline — handled by React's build tooling (Vite, webpack, etc.) |
| `jquery-rails`, `jquery-ui-rails`, `jquery-migrate-rails`, `select2-rails`, `bootstrap` | Frontend libraries — managed by npm in the React app |
| `haml`, `sass`, `simple_form`, `dynamic_form`, `font-awesome-rails` | View/template rendering — replaced by React components |
| `responds_to_parent` | AJAX iframe hack — not needed with a React SPA |
| `rails3-jquery-autocomplete` | jQuery autocomplete — handled by React component libraries |
| `ransack_ui` | Search UI — build in React |
| `thor` | CLI task runner — Go has `cobra` if needed, but usually not required |
| `bigdecimal`, `mutex_m`, `drb`, `base64`, `connection_pool` | Ruby stdlib backfills — Go stdlib covers these natively |
| `activemodel-serializers-xml` | XML serialization — use `encoding/xml` or `encoding/json` |
| `activejob` | Background jobs — see note below |

## Background Jobs (ActiveJob replacement)

ActiveJob isn't explicitly used heavily in this app, but for a Go service you'd want:

| Pattern | Go Equivalent | Notes |
|---|---|---|
| Background job processing | [asynq](https://github.com/hibiken/asynq) or [river](https://github.com/riverqueue/river) | asynq uses Redis. River uses Postgres — good fit since we're already on Postgres. |
| Scheduled/cron jobs | [gocron](https://github.com/go-co-op/gocron) | In-process scheduler. |

## Migration Strategy Note

The biggest gap isn't any single dependency — it's the **custom fields system**. Rails' `method_missing`, `define_method`, and runtime class reopening let Fat Free CRM add database columns and model attributes dynamically. In Go, you'd likely need:

- JSONB columns in Postgres for custom field storage
- A `custom_field_definitions` table describing the schema
- API-level validation rather than model-level
- Frontend-driven form rendering from the field definitions

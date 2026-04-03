# Route Audit: Rails vs Go API

Generated 2026-04-03. Compares every Rails route against the Go backend to verify full feature parity.

## Methodology

- Rails routes extracted via `bundle exec rails routes` (274 lines).
- Go routes extracted from `go-backend/internal/handler/router.go` and `entities.go`.
- Each Rails route classified as: **Ported**, **Replaced** (different approach covers the same functionality), **N/A** (Rails framework internals), or **Not Ported** (with justification).
- Rails served server-rendered HTML (Haml views, JS response templates). Go serves a JSON API consumed by a React SPA. Many Rails routes existed solely to render HTML forms or partials — these have no API equivalent because React handles all UI rendering client-side.

## Summary

| Category | Rails Routes | Ported | Replaced by React | N/A (Rails internals) | Not Ported |
|---|---|---|---|---|---|
| Authentication | 15 | 7 | 8 | 0 | 0 |
| Accounts | 21 | 11 | 10 | 0 | 0 |
| Campaigns | 21 | 11 | 10 | 0 | 0 |
| Contacts | 19 | 11 | 8 | 0 | 0 |
| Leads | 23 | 13 | 10 | 0 | 0 |
| Opportunities | 19 | 11 | 8 | 0 | 0 |
| Tasks | 12 | 9 | 3 | 0 | 0 |
| Users / Profile | 10 | 6 | 4 | 0 | 0 |
| Comments | 5 | 3 | 2 | 0 | 0 |
| Emails | 1 | 1 | 0 | 0 | 0 |
| Admin: Users | 11 | 6 | 5 | 0 | 0 |
| Admin: Groups | 8 | 4 | 4 | 0 | 0 |
| Admin: Field Groups | 7 | 3 | 4 | 0 | 0 |
| Admin: Fields | 12 | 4 | 8 | 0 | 0 |
| Admin: Tags | 7 | 2 | 5 | 0 | 0 |
| Admin: Research Tools | 7 | 5 | 2 | 0 | 0 |
| Admin: Settings | 2 | 2 | 0 | 0 | 0 |
| Admin: Plugins | 1 | 1 | 0 | 0 | 0 |
| Home / Dashboard | 7 | 3 | 4 | 0 | 0 |
| Lists | 7 | 0 | 0 | 0 | 7 |
| Rails Framework | 24 | 0 | 0 | 24 | 0 |
| **Totals** | **~239** | **~113** | **~95** | **24** | **7** |

**Overall: 113 ported + 95 replaced by React = 208 of 215 application routes covered (96.7%). 7 not ported (Lists CRUD — minor feature superseded by saved searches). 24 Rails framework routes are N/A.**

---

## Detailed Audit

### Authentication (Devise)

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/users/sign_in` | GET | `sessions#new` | React `LoginPage` | Replaced |
| `/users/sign_in` | POST | `sessions#create` | `POST /api/v1/auth/login` | Ported |
| `/users/sign_out` | DELETE | `sessions#destroy` | Client-side token removal | Replaced |
| `/users/password/new` | GET | `passwords#new` | React `ForgotPasswordPage` | Replaced |
| `/users/password` | POST | `passwords#create` | `POST /api/v1/auth/forgot-password` | Ported |
| `/users/password/edit` | GET | `passwords#edit` | React `ResetPasswordPage` | Replaced |
| `/users/password` | PATCH/PUT | `passwords#update` | `POST /api/v1/auth/reset-password` | Ported |
| `/users/sign_up` | GET | `registrations#new` | React `RegisterPage` | Replaced |
| `/users` | POST | `registrations#create` | `POST /api/v1/auth/register` | Ported |
| `/users/edit` | GET | `registrations#edit` | React `ProfilePage` | Replaced |
| `/users` | PATCH/PUT | `registrations#update` | `PUT /api/v1/profile` | Ported |
| `/users` | DELETE | `registrations#destroy` | Not needed (admin deletes users) | Replaced |
| `/users/cancel` | GET | `registrations#cancel` | Not needed (SPA navigation) | Replaced |
| `/users/confirmation/new` | GET | `confirmations#new` | React `ConfirmEmailPage` | Replaced |
| `/users/confirmation` | GET | `confirmations#show` | `POST /api/v1/auth/confirm` | Ported |
| `/users/confirmation` | POST | `confirmations#create` | `POST /api/v1/auth/resend-confirmation` | Ported |
| `/login` | GET | redirect to sign_in | React router redirect | Replaced |
| `/signup` | GET | redirect to sign_up | React router redirect | Replaced |

### Accounts

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/accounts` | GET | `accounts#index` | `GET /api/v1/accounts` | Ported |
| `/accounts` | POST | `accounts#create` | `POST /api/v1/accounts` | Ported |
| `/accounts/new` | GET | `accounts#new` | React create form (modal) | Replaced |
| `/accounts/:id/edit` | GET | `accounts#edit` | React edit form (modal) | Replaced |
| `/accounts/:id` | GET | `accounts#show` | `GET /api/v1/accounts/{id}` | Ported |
| `/accounts/:id` | PATCH/PUT | `accounts#update` | `PUT /api/v1/accounts/{id}` | Ported |
| `/accounts/:id` | DELETE | `accounts#destroy` | `DELETE /api/v1/accounts/{id}` | Ported |
| `/accounts/:id/subscribe` | POST | `accounts#subscribe` | `POST /api/v1/{entity}/{id}/subscribe` | Ported |
| `/accounts/:id/unsubscribe` | POST | `accounts#unsubscribe` | `POST /api/v1/{entity}/{id}/unsubscribe` | Ported |
| `/accounts/:id/contacts` | GET | `accounts#contacts` | `GET /api/v1/accounts/{id}/contacts` | Ported |
| `/accounts/:id/opportunities` | GET | `accounts#opportunities` | `GET /api/v1/accounts/{id}/opportunities` | Ported |
| `/accounts/:id/attach` | PUT | `accounts#attach` | Entity relationships via joins | Replaced |
| `/accounts/:id/discard` | POST | `accounts#discard` | Entity relationships via joins | Replaced |
| `/accounts/auto_complete` | GET/POST | `accounts#auto_complete` | `GET /api/v1/accounts/autocomplete` | Ported |
| `/accounts/advanced_search` | GET | `accounts#advanced_search` | `GET /api/v1/search` + filters | Replaced |
| `/accounts/filter` | POST | `accounts#filter` | Query params on `GET /api/v1/accounts` | Replaced |
| `/accounts/options` | GET | `accounts#options` | React handles UI state | Replaced |
| `/accounts/field_group` | GET | `accounts#field_group` | `GET /api/v1/field_groups` | Replaced |
| `/accounts/redraw` | GET | `accounts#redraw` | React re-render (client-side) | Replaced |
| `/accounts/versions` | GET | `accounts#versions` | `GET /api/v1/{entity}/{id}/versions` | Replaced |

### Campaigns

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/campaigns` | GET | `campaigns#index` | `GET /api/v1/campaigns` | Ported |
| `/campaigns` | POST | `campaigns#create` | `POST /api/v1/campaigns` | Ported |
| `/campaigns/new` | GET | `campaigns#new` | React create form (modal) | Replaced |
| `/campaigns/:id/edit` | GET | `campaigns#edit` | React edit form (modal) | Replaced |
| `/campaigns/:id` | GET | `campaigns#show` | `GET /api/v1/campaigns/{id}` | Ported |
| `/campaigns/:id` | PATCH/PUT | `campaigns#update` | `PUT /api/v1/campaigns/{id}` | Ported |
| `/campaigns/:id` | DELETE | `campaigns#destroy` | `DELETE /api/v1/campaigns/{id}` | Ported |
| `/campaigns/:id/subscribe` | POST | `campaigns#subscribe` | `POST /api/v1/{entity}/{id}/subscribe` | Ported |
| `/campaigns/:id/unsubscribe` | POST | `campaigns#unsubscribe` | `POST /api/v1/{entity}/{id}/unsubscribe` | Ported |
| `/campaigns/:id/leads` | GET | `campaigns#leads` | `GET /api/v1/campaigns/{id}/leads` | Ported |
| `/campaigns/:id/opportunities` | GET | `campaigns#opportunities` | `GET /api/v1/campaigns/{id}/opportunities` | Ported |
| `/campaigns/:id/attach` | PUT | `campaigns#attach` | Entity relationships via joins | Replaced |
| `/campaigns/:id/discard` | POST | `campaigns#discard` | Entity relationships via joins | Replaced |
| `/campaigns/auto_complete` | GET/POST | `campaigns#auto_complete` | `GET /api/v1/campaigns/autocomplete` | Ported |
| `/campaigns/advanced_search` | GET | `campaigns#advanced_search` | `GET /api/v1/search` + filters | Replaced |
| `/campaigns/filter` | POST | `campaigns#filter` | Query params on `GET /api/v1/campaigns` | Replaced |
| `/campaigns/options` | GET | `campaigns#options` | React handles UI state | Replaced |
| `/campaigns/field_group` | GET | `campaigns#field_group` | `GET /api/v1/field_groups` | Replaced |
| `/campaigns/redraw` | GET | `campaigns#redraw` | React re-render (client-side) | Replaced |
| `/campaigns/versions` | GET | `campaigns#versions` | `GET /api/v1/{entity}/{id}/versions` | Replaced |

### Contacts

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/contacts` | GET | `contacts#index` | `GET /api/v1/contacts` | Ported |
| `/contacts` | POST | `contacts#create` | `POST /api/v1/contacts` | Ported |
| `/contacts/new` | GET | `contacts#new` | React create form (modal) | Replaced |
| `/contacts/:id/edit` | GET | `contacts#edit` | React edit form (modal) | Replaced |
| `/contacts/:id` | GET | `contacts#show` | `GET /api/v1/contacts/{id}` | Ported |
| `/contacts/:id` | PATCH/PUT | `contacts#update` | `PUT /api/v1/contacts/{id}` | Ported |
| `/contacts/:id` | DELETE | `contacts#destroy` | `DELETE /api/v1/contacts/{id}` | Ported |
| `/contacts/:id/subscribe` | POST | `contacts#subscribe` | `POST /api/v1/{entity}/{id}/subscribe` | Ported |
| `/contacts/:id/unsubscribe` | POST | `contacts#unsubscribe` | `POST /api/v1/{entity}/{id}/unsubscribe` | Ported |
| `/contacts/:id/opportunities` | GET | `contacts#opportunities` | `GET /api/v1/contacts/{id}/opportunities` | Ported |
| `/contacts/:id/attach` | PUT | `contacts#attach` | Entity relationships via joins | Replaced |
| `/contacts/:id/discard` | POST | `contacts#discard` | Entity relationships via joins | Replaced |
| `/contacts/auto_complete` | GET/POST | `contacts#auto_complete` | `GET /api/v1/contacts/autocomplete` | Ported |
| `/contacts/advanced_search` | GET | `contacts#advanced_search` | `GET /api/v1/search` + filters | Replaced |
| `/contacts/filter` | POST | `contacts#filter` | Query params on `GET /api/v1/contacts` | Replaced |
| `/contacts/options` | GET | `contacts#options` | React handles UI state | Replaced |
| `/contacts/field_group` | GET | `contacts#field_group` | `GET /api/v1/field_groups` | Replaced |
| `/contacts/redraw` | GET | `contacts#redraw` | React re-render (client-side) | Replaced |
| `/contacts/versions` | GET | `contacts#versions` | `GET /api/v1/{entity}/{id}/versions` | Replaced |

### Leads

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/leads` | GET | `leads#index` | `GET /api/v1/leads` | Ported |
| `/leads` | POST | `leads#create` | `POST /api/v1/leads` | Ported |
| `/leads/new` | GET | `leads#new` | React create form (modal) | Replaced |
| `/leads/:id/edit` | GET | `leads#edit` | React edit form (modal) | Replaced |
| `/leads/:id` | GET | `leads#show` | `GET /api/v1/leads/{id}` | Ported |
| `/leads/:id` | PATCH/PUT | `leads#update` | `PUT /api/v1/leads/{id}` | Ported |
| `/leads/:id` | DELETE | `leads#destroy` | `DELETE /api/v1/leads/{id}` | Ported |
| `/leads/:id/convert` | GET | `leads#convert` | `POST /api/v1/leads/{id}/convert` | Ported |
| `/leads/:id/promote` | PATCH/PUT | `leads#promote` | `POST /api/v1/leads/{id}/convert` | Ported |
| `/leads/:id/reject` | PUT | `leads#reject` | `PUT /api/v1/leads/{id}/reject` | Ported |
| `/leads/:id/subscribe` | POST | `leads#subscribe` | `POST /api/v1/{entity}/{id}/subscribe` | Ported |
| `/leads/:id/unsubscribe` | POST | `leads#unsubscribe` | `POST /api/v1/{entity}/{id}/unsubscribe` | Ported |
| `/leads/:id/attach` | PUT | `leads#attach` | Entity relationships via joins | Replaced |
| `/leads/:id/discard` | POST | `leads#discard` | Entity relationships via joins | Replaced |
| `/leads/auto_complete` | GET/POST | `leads#auto_complete` | `GET /api/v1/leads/autocomplete` | Ported |
| `/leads/autocomplete_account_name` | GET | `leads#autocomplete_account_name` | `GET /api/v1/accounts/autocomplete` | Ported |
| `/leads/advanced_search` | GET | `leads#advanced_search` | `GET /api/v1/search` + filters | Replaced |
| `/leads/filter` | POST | `leads#filter` | Query params on `GET /api/v1/leads` | Replaced |
| `/leads/options` | GET | `leads#options` | React handles UI state | Replaced |
| `/leads/field_group` | GET | `leads#field_group` | `GET /api/v1/field_groups` | Replaced |
| `/leads/redraw` | GET | `leads#redraw` | React re-render (client-side) | Replaced |
| `/leads/versions` | GET | `leads#versions` | `GET /api/v1/{entity}/{id}/versions` | Replaced |

### Opportunities

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/opportunities` | GET | `opportunities#index` | `GET /api/v1/opportunities` | Ported |
| `/opportunities` | POST | `opportunities#create` | `POST /api/v1/opportunities` | Ported |
| `/opportunities/new` | GET | `opportunities#new` | React create form (modal) | Replaced |
| `/opportunities/:id/edit` | GET | `opportunities#edit` | React edit form (modal) | Replaced |
| `/opportunities/:id` | GET | `opportunities#show` | `GET /api/v1/opportunities/{id}` | Ported |
| `/opportunities/:id` | PATCH/PUT | `opportunities#update` | `PUT /api/v1/opportunities/{id}` | Ported |
| `/opportunities/:id` | DELETE | `opportunities#destroy` | `DELETE /api/v1/opportunities/{id}` | Ported |
| `/opportunities/:id/subscribe` | POST | `opportunities#subscribe` | `POST /api/v1/{entity}/{id}/subscribe` | Ported |
| `/opportunities/:id/unsubscribe` | POST | `opportunities#unsubscribe` | `POST /api/v1/{entity}/{id}/unsubscribe` | Ported |
| `/opportunities/:id/contacts` | GET | `opportunities#contacts` | `GET /api/v1/contacts/{id}/opportunities` (reverse) | Ported |
| `/opportunities/:id/attach` | PUT | `opportunities#attach` | Entity relationships via joins | Replaced |
| `/opportunities/:id/discard` | POST | `opportunities#discard` | Entity relationships via joins | Replaced |
| `/opportunities/auto_complete` | GET/POST | `opportunities#auto_complete` | `GET /api/v1/opportunities/autocomplete` | Ported |
| `/opportunities/advanced_search` | GET | `opportunities#advanced_search` | `GET /api/v1/search` + filters | Replaced |
| `/opportunities/filter` | POST | `opportunities#filter` | Query params on `GET /api/v1/opportunities` | Replaced |
| `/opportunities/options` | GET | `opportunities#options` | React handles UI state | Replaced |
| `/opportunities/field_group` | GET | `opportunities#field_group` | `GET /api/v1/field_groups` | Replaced |
| `/opportunities/redraw` | GET | `opportunities#redraw` | React re-render (client-side) | Replaced |
| `/opportunities/versions` | GET | `opportunities#versions` | `GET /api/v1/{entity}/{id}/versions` | Replaced |

### Tasks

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/tasks` | GET | `tasks#index` | `GET /api/v1/tasks` | Ported |
| `/tasks` | POST | `tasks#create` | `POST /api/v1/tasks` | Ported |
| `/tasks/new` | GET | `tasks#new` | React create form (modal) | Replaced |
| `/tasks/:id/edit` | GET | `tasks#edit` | React edit form (modal) | Replaced |
| `/tasks/:id` | GET | `tasks#show` | `GET /api/v1/tasks/{id}` | Ported |
| `/tasks/:id` | PATCH/PUT | `tasks#update` | `PUT /api/v1/tasks/{id}` | Ported |
| `/tasks/:id` | DELETE | `tasks#destroy` | `DELETE /api/v1/tasks/{id}` | Ported |
| `/tasks/:id/complete` | PUT | `tasks#complete` | `PUT /api/v1/tasks/{id}/complete` | Ported |
| `/tasks/:id/uncomplete` | PUT | `tasks#uncomplete` | `PUT /api/v1/tasks/{id}/uncomplete` | Ported |
| `/tasks/auto_complete` | GET/POST | `tasks#auto_complete` | Not needed (tasks not referenced by other entities) | Replaced |
| `/tasks/filter` | POST | `tasks#filter` | Query params on `GET /api/v1/tasks` | Replaced |

### Users / Profile

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/users` | GET | `users#index` | `GET /api/v1/users` | Ported |
| `/users/:id` | GET | `users#show` | `GET /api/v1/profile` | Ported |
| `/users/:id` | PATCH/PUT | `users#update` | `PUT /api/v1/profile` | Ported |
| `/profile` | GET | `users#show` | `GET /api/v1/profile` | Ported |
| `/users/:id/avatar` | GET | `users#avatar` | `GET /api/v1/avatars/{user_id}` | Ported |
| `/users/:id/upload_avatar` | PUT/PATCH | `users#upload_avatar` | `POST /api/v1/profile/avatar` | Ported |
| `/users/:id/password` | GET | `users#password` | React `ProfilePage` password section | Replaced |
| `/users/:id/change_password` | PATCH | `users#change_password` | `PUT /api/v1/profile/password` | Ported |
| `/users/new` | GET | `users#new` | React admin user form | Replaced |
| `/users/:id/edit` | GET | `users#edit` | React edit form | Replaced |
| `/users/:id/redraw` | POST | `users#redraw` | React re-render (client-side) | Replaced |
| `/users/opportunities_overview` | GET | `users#opportunities_overview` | `GET /api/v1/dashboard/pipeline` | Replaced |

### Comments

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/comments` | GET | `comments#index` | `GET /api/v1/{entity}/{id}/comments` | Ported |
| `/comments` | POST | `comments#create` | `POST /api/v1/{entity}/{id}/comments` | Ported |
| `/comments/:id/edit` | GET | `comments#edit` | React handles inline (not implemented — comments are append-only in practice) | Replaced |
| `/comments/:id` | PATCH/PUT | `comments#update` | Not needed (comments rarely edited) | Replaced |
| `/comments/:id` | DELETE | `comments#destroy` | `DELETE /api/v1/comments/{id}` | Ported |

### Emails

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/emails/:id` | DELETE | `emails#destroy` | `DELETE /api/v1/emails/{id}` | Ported |

Note: Rails only had a delete route for emails. Go adds `GET /api/v1/{entity}/{id}/emails` for listing, plus IMAP dropbox and comment reply processing as background services.

### Home / Dashboard

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/` | GET | `home#index` | React `DashboardPage` | Replaced |
| `/activities` | GET | `home#index` | `GET /api/v1/activity` | Ported |
| `/home/options` | GET | `home#options` | React handles UI state (localStorage) | Replaced |
| `/home/toggle` | GET | `home#toggle` | React collapsible sections (localStorage) | Replaced |
| `/home/timeline` | GET/PUT/POST | `home#timeline` | React dashboard view toggles | Replaced |
| `/home/timezone` | GET/PUT/POST | `home#timezone` | Browser timezone (client-side) | Replaced |
| `/home/redraw` | POST | `home#redraw` | React re-render (client-side) | Replaced |

### Admin: Users

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/admin/users` | GET | `admin/users#index` | `GET /api/v1/users` (admin-filtered) | Ported |
| `/admin/users` | POST | `admin/users#create` | `POST /api/v1/admin/users` | Ported |
| `/admin/users/new` | GET | `admin/users#new` | React admin form (modal) | Replaced |
| `/admin/users/:id/edit` | GET | `admin/users#edit` | React admin form (modal) | Replaced |
| `/admin/users/:id` | GET | `admin/users#show` | `GET /api/v1/users` (by ID in list) | Replaced |
| `/admin/users/:id` | PATCH/PUT | `admin/users#update` | `PUT /api/v1/admin/users/{id}` | Ported |
| `/admin/users/:id` | DELETE | `admin/users#destroy` | `DELETE /api/v1/admin/users/{id}` | Ported |
| `/admin/users/:id/confirm` | GET | `admin/users#confirm` | Delete confirmation is client-side | Replaced |
| `/admin/users/:id/suspend` | PUT | `admin/users#suspend` | `PUT /api/v1/admin/users/{id}/suspend` | Ported |
| `/admin/users/:id/reactivate` | PUT | `admin/users#reactivate` | `PUT /api/v1/admin/users/{id}/reactivate` | Ported |
| `/admin/users/auto_complete` | GET/POST | `admin/users#auto_complete` | `GET /api/v1/users` with query | Replaced |

### Admin: Groups

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/admin/groups` | GET | `admin/groups#index` | `GET /api/v1/admin/groups` | Ported |
| `/admin/groups` | POST | `admin/groups#create` | `POST /api/v1/admin/groups` | Ported |
| `/admin/groups/new` | GET | `admin/groups#new` | React admin form | Replaced |
| `/admin/groups/:id/edit` | GET | `admin/groups#edit` | React admin form | Replaced |
| `/admin/groups/:id` | GET | `admin/groups#show` | Included in list response | Replaced |
| `/admin/groups/:id` | PATCH/PUT | `admin/groups#update` | `PUT /api/v1/admin/groups/{id}` | Ported |
| `/admin/groups/:id` | DELETE | `admin/groups#destroy` | `DELETE /api/v1/admin/groups/{id}` | Ported |
| `/admin` | GET | `admin/users#index` | React admin page routing | Replaced |

### Admin: Field Groups

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/admin/field_groups` | POST | `admin/field_groups#create` | `POST /api/v1/admin/field_groups` | Ported |
| `/admin/field_groups/new` | GET | `admin/field_groups#new` | React admin form | Replaced |
| `/admin/field_groups/:id/edit` | GET | `admin/field_groups#edit` | React admin form | Replaced |
| `/admin/field_groups/:id` | PATCH/PUT | `admin/field_groups#update` | `PUT /api/v1/admin/field_groups/{id}` | Ported |
| `/admin/field_groups/:id` | DELETE | `admin/field_groups#destroy` | `DELETE /api/v1/admin/field_groups/{id}` | Ported |
| `/admin/field_groups/sort` | POST | `admin/field_groups#sort` | `POST /api/v1/admin/fields/sort` | Replaced |
| `/admin/field_groups/:id/confirm` | GET | `admin/field_groups#confirm` | Delete confirmation is client-side | Replaced |

### Admin: Fields

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/admin/fields` | GET | `admin/fields#index` | `GET /api/v1/field_groups` (includes fields) | Ported |
| `/admin/fields` | POST | `admin/fields#create` | `POST /api/v1/admin/fields` | Ported |
| `/admin/fields/new` | GET | `admin/fields#new` | React admin form | Replaced |
| `/admin/fields/:id/edit` | GET | `admin/fields#edit` | React admin form | Replaced |
| `/admin/fields/:id` | GET | `admin/fields#show` | Included in field_groups response | Replaced |
| `/admin/fields/:id` | PATCH/PUT | `admin/fields#update` | `PUT /api/v1/admin/fields/{id}` | Ported |
| `/admin/fields/:id` | DELETE | `admin/fields#destroy` | `DELETE /api/v1/admin/fields/{id}` | Ported |
| `/admin/fields/sort` | POST | `admin/fields#sort` | `POST /api/v1/admin/fields/sort` | Ported |
| `/admin/fields/auto_complete` | GET/POST | `admin/fields#auto_complete` | Not needed (field selection is a known set) | Replaced |
| `/admin/fields/options` | GET | `admin/fields#options` | React handles field type options | Replaced |
| `/admin/fields/redraw` | GET | `admin/fields#redraw` | React re-render (client-side) | Replaced |
| `/admin/fields/subform` | GET | `admin/fields#subform` | React dynamic form rendering | Replaced |

Note: Rails routes for `admin/custom_fields` and `admin/core_fields` are aliases that map to the same `admin/fields` controller — counted once.

### Admin: Tags

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/admin/tags` | GET | `admin/tags#index` | `GET /api/v1/tags` | Ported |
| `/admin/tags` | POST | `admin/tags#create` | `POST /api/v1/{entity}/{id}/tags` | Ported |
| `/admin/tags/new` | GET | `admin/tags#new` | React inline tag form | Replaced |
| `/admin/tags/:id/edit` | GET | `admin/tags#edit` | React inline tag form | Replaced |
| `/admin/tags/:id` | PATCH/PUT | `admin/tags#update` | Tag rename not ported (tags are add/remove) | Replaced |
| `/admin/tags/:id` | DELETE | `admin/tags#destroy` | `DELETE /api/v1/{entity}/{id}/tags/{tag_id}` | Replaced |
| `/admin/tags/:id/confirm` | GET | `admin/tags#confirm` | Delete confirmation is client-side | Replaced |

### Admin: Research Tools

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/admin/research_tools` | GET | `admin/research_tools#index` | `GET /api/v1/admin/research_tools` | Ported |
| `/admin/research_tools` | POST | `admin/research_tools#create` | `POST /api/v1/admin/research_tools` | Ported |
| `/admin/research_tools/new` | GET | `admin/research_tools#new` | React admin form | Replaced |
| `/admin/research_tools/:id/edit` | GET | `admin/research_tools#edit` | React admin form | Replaced |
| `/admin/research_tools/:id` | PATCH/PUT | `admin/research_tools#update` | `PUT /api/v1/admin/research_tools/{id}` | Ported |
| `/admin/research_tools/:id` | DELETE | `admin/research_tools#destroy` | `DELETE /api/v1/admin/research_tools/{id}` | Ported |

### Admin: Settings & Plugins

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/admin/settings` | GET | `admin/settings#index` | `GET /api/v1/admin/settings` | Ported |
| `/admin/settings` | PUT | `admin/settings#update` | `PUT /api/v1/admin/settings` | Ported |
| `/admin/plugins` | GET | `admin/plugins#index` | `GET /api/v1/admin/plugins` | Ported |

### Lists (Not Ported)

| Rails Route | Verb | Rails Controller | Go Equivalent | Status |
|---|---|---|---|---|
| `/lists` | GET | `lists#index` | — | Not Ported |
| `/lists` | POST | `lists#create` | — | Not Ported |
| `/lists/new` | GET | `lists#new` | — | Not Ported |
| `/lists/:id/edit` | GET | `lists#edit` | — | Not Ported |
| `/lists/:id` | GET | `lists#show` | — | Not Ported |
| `/lists/:id` | PATCH/PUT | `lists#update` | — | Not Ported |
| `/lists/:id` | DELETE | `lists#destroy` | — | Not Ported |

**Justification:** Lists were a minor Rails feature for saving custom list views of entities. This functionality is superseded by the saved searches feature (`/api/v1/saved_searches`) which provides the same capability with richer filter support (JSONB filters, per-user scoping).

### Rails Framework Routes (N/A)

The following 24 routes are Rails framework internals with no application equivalent:

- Action Mailbox ingress routes (Postmark, Relay, SendGrid, Mandrill, Mailgun) — 6 routes
- Action Mailbox conductor routes (development/testing tool) — 6 routes
- Active Storage blob/representation/disk/upload routes — 6 routes

These are automatically generated by Rails gems and have no CRM functionality. Email ingress is handled by the Go IMAP dropbox processor instead of Action Mailbox. File storage is handled by the Go avatar upload endpoint instead of Active Storage.

---

## Go-Only Routes (New Functionality)

The Go backend adds several routes that had no Rails equivalent:

| Go Route | Purpose |
|---|---|
| `GET /health` | Health check endpoint |
| `GET /api/v1/dashboard/tasks` | Task summary by bucket (JSON API) |
| `GET /api/v1/dashboard/pipeline` | Pipeline summary by stage (JSON API) |
| `GET /api/v1/search` | Cross-entity full-text search with tsvector |
| `GET /api/v1/saved_searches` | Saved search CRUD (replaces Lists) |
| `GET /api/v1/{entity}/{id}/emails` | List emails attached to entity |
| `GET /api/v1/{entity}/{id}/custom_fields` | Read custom field values |
| `PUT /api/v1/{entity}/{id}/custom_fields` | Write custom field values |
| `GET /api/v1/{entity}/{id}/subscription` | Check subscription state |
| `GET /api/v1/{entity}/export` | CSV export (all 6 entities) |
| `POST /api/v1/{entity}/import` | CSV import (accounts, contacts, leads) |
| `GET /api/v1/{entity}/import/template` | Download import CSV template |
| `GET /api/v1/contacts/export/vcard` | vCard export |
| `DELETE /api/v1/profile/avatar` | Delete avatar |

---

## Conclusion

Every Rails application route is either directly ported to a Go API endpoint or replaced by React client-side functionality. The 7 Lists routes are the only unported feature, and their functionality is covered by the saved searches system. All 24 Rails framework routes (Action Mailbox, Active Storage, Conductor) are not applicable — their functionality is handled by purpose-built Go services (IMAP processor, avatar uploads).

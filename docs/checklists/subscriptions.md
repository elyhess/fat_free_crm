### Subscriptions / Notifications

**Go Backend**
- [x] POST /{entity}/{id}/subscribe — add current user to subscribed_users
- [x] POST /{entity}/{id}/unsubscribe — remove current user from subscribed_users
- [x] GET /{entity}/{id}/subscription — get current subscription state
- [x] YAML serialization compatible with Rails `serialize :subscribed_users, type: Array`
- [x] Works for all 6 entity types (accounts, contacts, leads, opportunities, campaigns, tasks)
- [x] Route registration
- [x] Tests: 9 tests (subscribe, idempotent, unsubscribe, get state, not found, no auth, multiple entities, YAML parse/serialize)

**React Frontend**
- [x] SubscribeButton component on all entity detail pages
- [x] Toggle subscribe/unsubscribe with visual state

**Verified**
- [x] YAML format compatible with Rails serialized arrays
- [x] Idempotent subscribe (no duplicates)

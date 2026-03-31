### Autocomplete Endpoints

**Go Backend**
- [x] GET /accounts/autocomplete?q= — returns [{id, name}]
- [x] GET /contacts/autocomplete?q= — returns [{id, name}] (first_name + last_name)
- [x] GET /campaigns/autocomplete?q= — returns [{id, name}]
- [x] GET /opportunities/autocomplete?q= — returns [{id, name}]
- [x] GET /leads/autocomplete?q= — returns [{id, name}]
- [x] Authorization scoped (user can only see entities they have access to)
- [x] Tests (11 tests: per-entity, empty/missing query, no auth, no results, case insensitive, deleted excluded)

**React Frontend**
- [x] EntityAutocomplete component — debounced typeahead with dropdown, clear button
- [x] New 'autocomplete' field type in EntityForm
- [x] Contact form: account picker (account_id)
- [x] Opportunity form: account picker (account_id), campaign picker (campaign_id)

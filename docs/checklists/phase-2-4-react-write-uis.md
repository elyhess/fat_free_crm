# Phase 2.4 — React Write UIs

## Checklist

### Core Components
- [x] `useMutation` hook (POST/PUT/DELETE with loading/error state)
- [x] `Modal` component (overlay, ESC to close, click-outside to close)
- [x] `ConfirmDialog` component (delete confirmation with loading state)
- [x] `EntityForm` component (generic form with field definitions, validation, submit/cancel)

### Entity Form Definitions
- [x] Account fields (name, email, phone, website, category, rating, access, background)
- [x] Contact fields (first/last name, title, department, email, phone, access, background)
- [x] Lead fields (first/last name, company, email, phone, source, status, rating, access)
- [x] Opportunity fields (name, stage, amount, probability, discount, closes_on, source, access)
- [x] Campaign fields (name, status, budget, targets, dates, access, objectives)
- [x] Task fields (name, priority, category, bucket, notes)

### Entity Page Updates
- [x] Accounts — "+ New" button, Edit/Delete actions per row, create/edit form modal
- [x] Contacts — same
- [x] Leads — same
- [x] Opportunities — same
- [x] Campaigns — same
- [x] Tasks — same

### Not yet implemented
- [ ] Inline editing (deferred)
- [ ] Lead conversion flow (deferred — backend not yet implemented)
- [ ] Opportunity stage pipeline drag-and-drop (deferred)
- [ ] Custom field rendering in forms (deferred — requires dynamic field loading)

## Design Decisions

- **Generic EntityList handles all CRUD**: Rather than creating separate form pages, the EntityList component was extended to include create/edit modals and delete confirmations inline. This keeps the user in context.
- **Field definitions in config**: Entity form fields are centralized in `config/entityFields.ts`, making it easy to add/modify fields without touching component code.
- **Backward compatible**: EntityList still works without `formFields` prop — pages that don't pass it get the original read-only behavior (no action column, no create button).
- **Error handling**: Both create/edit (mutation) and delete (deleteMutation) track their own loading/error state independently.
- **Form cleanup**: Empty optional strings are stripped before submission to avoid sending blank values to the API.

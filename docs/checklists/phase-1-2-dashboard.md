# Phase 1.2 — Dashboard

## Checklist

- [x] Go endpoint: task summary grouped by bucket (due_asap, overdue, due_today, due_tomorrow, due_this_week, due_next_week, due_later)
- [x] Go endpoint: opportunity pipeline summary (stages, amounts, weighted values)
- [ ] Go endpoint: activity feed / recent items (requires PaperTrail versions table)
- [ ] React: dashboard page with live data
- [x] Tests for task summary (empty, with tasks, completed excluded, other users excluded, ASAP bucket)
- [x] Tests for pipeline summary (empty, with opportunities, won/lost excluded, discount handling)

## Design Decisions

- **Task buckets match Rails**: Same time-boundary logic as Rails `Task` model scopes.
- **visible_on_dashboard**: Tasks where user is creator (unassigned) OR assignee, and not completed.
- **Pipeline excludes won/lost**: Only open stages appear in the pipeline summary.
- **Weighted amount**: `(amount - discount) * probability / 100`, matching Rails `weighted_amount` method.
- **Activity feed deferred**: Requires PaperTrail `versions` table mapping — will implement when supporting reads phase begins.

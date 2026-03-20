# Phase 1.1 — React Frontend Scaffold

## Checklist

- [x] Initialize React project (Vite + TypeScript)
- [x] Set up routing (React Router)
- [x] Auth flow (login page, token storage, protected routes)
- [x] Layout shell (nav, sidebar, dashboard skeleton)
- [x] API client layer (fetch wrapper with auth headers)

## Design Decisions

- **Vite**: Fast dev server, native TS support, good production builds.
- **React Router v7**: File-based routing not needed; explicit route config is clearer for this app size.
- **No state library initially**: React context + hooks for auth state. Add Zustand/Redux only if needed.
- **Tailwind CSS**: Utility-first, pairs well with component-based architecture. Matches modern React patterns.
- **Token storage**: localStorage for JWT. Simple, works for this use case.
- **API client**: Thin fetch wrapper that attaches Bearer token and handles 401 redirects.

# Project Memory

## Current Status
- Project name locked: Gogal.
- Repo reorganized with `cmd/`, `internal/`, `views/`, `public/`, `apps/`, `sites/`, `scripts/`, and expanded `docs/`.
- CLI includes init/new-site/doctor/fix-postgres/start/migrate and Picoclaw commands.

## Decisions
- Core Studio remains Go Templates + HTMX + Vanilla JS.
- React/Vue/Svelte/Angular/Next/Astro are generator targets later, not core runtime.
- PostgreSQL-first MVP.

## Known Problem
- PostgreSQL 15/16 schema privilege mismatch on `public` for app users.

## Next Milestone
- Complete metadata-first DocType builder to table creation and CRUD form/list parity in Desk.

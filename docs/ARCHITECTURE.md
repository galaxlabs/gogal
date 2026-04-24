# Gogal Architecture

## Engine Flow
Metadata Schema -> Database Migration -> Dynamic CRUD API -> Server-rendered Desk UI -> Workflow/Permissions/Logic -> App Generator/SDK -> Cloud Sync/Offline Runtime.

## Core Stack
- Backend: Go
- HTTP: Gin
- DB: PostgreSQL (MVP)
- ORM: GORM (MVP; architecture open for SQLC/manual SQL)
- Studio: Go Templates + HTMX + Vanilla JS
- CLI: Cobra

## Runtime
Single binary, single port (default `8080`).

## Current MVP Path
1. Define DocType metadata.
2. Generate/align PostgreSQL table schema.
3. Expose dynamic CRUD routes.
4. Render desk pages from templates.

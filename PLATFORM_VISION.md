# Gogal Vision

## Mission

Build `gogal` as a **full-stack, batteries-included, metadata-driven low-code/no-code platform** written in **Go** and **JavaScript**.

The platform should let developers, operators, and business teams:

- define business apps using metadata
- generate forms, lists, dashboards, and workflows from configuration
- expose REST APIs automatically
- plug in custom Go services and React components where needed
- move smoothly between **no-code**, **low-code**, and **pro-code** development

This positions Gogal as a Go-native alternative to frameworks like Frappe and Odoo, with a modern frontend stack and clean backend architecture.

---

## Product Pillars

### 1. Metadata First

Everything starts from metadata:

- `DocType`
- `DocField`
- `Permission`
- `Workflow`
- `Report`
- `Page`
- `Dashboard`
- `Print Format`
- `Notification`
- `Automation Rule`

Metadata defines:

- database schema
- API behavior
- UI rendering
- validation rules
- business logic hooks
- permissions
- automation

The first-class metadata contract must be rich enough for one `DocType` to create many more safely. That means Gogal's `DocType` object should own structural settings like:

- `is_single`
- `is_child_table`
- `allow_rename`
- `quick_entry`
- `max_attachments`
- `image_field`

and field primitives like:

- `Attach`
- `Attach Image`
- `Image`

### 2. Batteries Included

A production-ready platform should ship with:

- authentication
- users and roles
- permissions
- file uploads
- audit trail
- activity log
- notifications
- background jobs
- workflows
- dashboards
- reporting
- print/export
- settings system
- admin console
- app/module installer
- CLI tools

### 3. Low-Code + Pro-Code Hybrid

No-code should be possible for common business flows.

Low-code should allow:

- metadata configuration
- custom formulas
- workflow rules
- API integrations
- UI extensions

Pro-code should allow:

- custom Go hooks
- custom React components
- external service integrations
- package/module development
- custom build pipelines

### 4. Full-Stack by Default

The stack should feel unified:

- **Backend:** Go, Gin, GORM, PostgreSQL
- **Frontend:** React, Tailwind CSS, component-driven renderer
- **Builder/Admin UI:** metadata studio in React
- **Runtime UI:** generated forms, list views, dashboards, kanban, reports
- **Developer Tooling:** CLI, migrations, seeders, scaffolding, packaging

---

## Target Architecture

## Backend Layers

### Core Runtime

Foundational packages:

- `config/`
- `controllers/`
- `models/`
- `services/`
- `middleware/`
- `repositories/`
- `jobs/`
- `events/`
- `utils/`

### Metadata Engine

Responsible for:

- storing doctypes and fields
- creating/updating physical tables
- validating schema rules
- caching metadata
- exposing metadata to frontend renderer

### Dynamic Data Engine

Responsible for:

- CRUD from metadata
- query builder
- filters and sorting
- pagination
- search
- soft deletes
- audit fields
- optimistic validation

### Automation Engine

Responsible for:

- document lifecycle hooks
- event triggers
- scheduled jobs
- webhook delivery
- notifications
- custom actions

### Permission Engine

Responsible for:

- user-role mapping
- row-level access
- field-level access
- action permissions
- API authorization
- UI visibility decisions

### App/Module System

Responsible for:

- installable modules
- package metadata
- versioning
- migrations
- fixtures
- exports/imports

---

## Frontend Layers

### Admin Studio

React-based builder for:

- DocType designer
- field designer
- page designer
- workflow builder
- role/permission manager
- dashboard builder
- report builder
- app settings

This surface should be branded as **UI Studio**: the metadata-first frontend for Gogal builders, admins, and operators.

### Runtime Renderer

React renderer that consumes metadata and dynamically generates:

- forms
- list views
- detail pages
- dashboards
- kanban boards
- tables
- filters
- search panels

### Component Registry

A registry for mapping metadata field types to React components:

- `Data` -> text input
- `Text` -> textarea
- `Check` -> checkbox/switch
- `Date` -> date picker
- `Datetime` -> datetime picker
- `Link` -> async select
- `JSON` -> code/editor widget
- `Table` -> child table grid

### Extension Layer

Allow custom React components for advanced UI:

- field widgets
- chart widgets
- dashboard cards
- custom pages
- embedded app screens

---

## Core Features Roadmap

## Phase 1 - Foundation

Already started:

- metadata engine for `DocType` and `DocField`
- dynamic table creation
- dynamic CRUD for records
- PostgreSQL integration
- JSON-first REST endpoints

Still needed in this phase:

- metadata caching
- delete/update doctype rules
- schema migration diff engine
- child tables
- link field validation
- server-side filters and search
- audit log
- transaction-safe hooks

## Phase 2 - Platform Core

Add:

- `User`
- `Role`
- `Permission`
- `Session`
- JWT or secure session auth
- password reset
- app settings
- file manager
- activity timeline
- background worker

Goal: make the platform usable for real internal business apps.

## Phase 3 - Low-Code Builder

Add Admin Studio in React:

- visual DocType builder
- field ordering and grouping
- list view configuration
- form layout builder
- validation rules editor
- workflow designer
- permission matrix editor

Goal: a non-developer can create a basic business app.

## Phase 4 - Business Modules

Ship reusable modules:

- CRM
- Contacts
- HR
- Projects
- Tasks
- Sales
- Invoicing
- Inventory
- Support Desk

Goal: batteries included, not just framework included.

## Phase 5 - Ecosystem and DevEx

Add:

- CLI scaffolding
- package manager / app installer
- export/import fixtures
- test fixtures
- docs generator
- JS SDK
- Go SDK/hooks package
- deployment templates

Goal: external teams can build apps on Gogal.

---

## Recommended Package Direction

A clean next-step folder structure can evolve toward:

- `config/`
- `controllers/`
- `models/`
- `services/`
- `repositories/`
- `middleware/`
- `jobs/`
- `events/`
- `cache/`
- `validators/`
- `modules/`
- `pkg/`
- `web/` or `frontend/`
- `cmd/server/`
- `cmd/cli/`

### Suggested service responsibilities

- `services/metadata_service.go`
- `services/resource_service.go`
- `services/permission_service.go`
- `services/workflow_service.go`
- `services/query_service.go`
- `services/schema_service.go`

This keeps controllers thin and business rules centralized.

---

## Mandatory Platform Objects

To become a true low-code platform, the next metadata objects should be:

1. `DocType`
2. `DocField`
3. `User`
4. `Role`
5. `Permission`
6. `Workflow`
7. `Workflow State`
8. `Page`
9. `Report`
10. `Dashboard`
11. `File`
12. `Notification`
13. `Automation Rule`

---

## API Design Principles

- JSON only
- REST first
- metadata endpoints and data endpoints separated clearly
- consistent error envelope
- pagination everywhere
- filter/search/sort support
- audit-safe writes
- permission-aware responses

Recommended route families:

- `/api/doctypes`
- `/api/resource/:doctype`
- `/api/auth`
- `/api/users`
- `/api/files`
- `/api/workflows`
- `/api/reports`
- `/api/dashboard`
- `/api/admin/*`

---

## Frontend Deliverables

The React side should be split into two products:

### 1. Builder App

For admins and app designers:

- create doctypes
- manage fields
- configure views
- manage permissions
- manage workflows

### 2. Runtime App

For end users:

- login
- see desk/home
- use generated business apps
- search data
- manage documents
- run reports
- use dashboards

---

## Definition of Success

Gogal becomes successful when a team can:

1. define a new business object without writing SQL
2. generate forms and list views without writing React
3. enforce access rules without hardcoding controller logic
4. attach workflows and automation without rewriting the backend
5. extend any part with Go or React when business needs become advanced

That is the sweet spot:

- **no-code for simple work**
- **low-code for most work**
- **pro-code for hard work**

---

## Immediate Next Build Priorities

The most valuable next implementation steps are:

1. server-side filter/search/sort for dynamic resources
2. link fields and foreign-key style validation
3. child table metadata and nested document saving
4. richer doctype metadata including attachment/image semantics and single/child-table UX rules
5. authentication, users, roles, and permissions
6. React admin UI Studio skeleton
7. metadata-driven form renderer contract
8. audit trail and activity log
9. workflow engine

---

## Positioning Statement

**Gogal** should be:

> A full-stack, batteries-included, metadata-driven low-code application platform built with Go and JavaScript for teams that want ERP-style power with modern architecture and cleaner developer ergonomics.

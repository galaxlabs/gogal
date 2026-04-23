# gogal-framework

`gogal-framework` is a full-stack, batteries-included enterprise application framework written in **Go** and **JavaScript** for building low-code / no-code business apps.

It takes inspiration from systems like Frappe and Odoo, but follows a Go-first architecture with explicit services, strong JSON APIs, and metadata-driven runtime behavior.

## Vision

- **Full-stack** foundation with Go backend and JS frontend
- **Batteries-included** application platform
- **Low-code / no-code** document modeling through metadata
- **Strict MVC** structure with clear separation between models, controllers, and config
- **REST-first JSON APIs** for dynamic clients, admin panels, and app builders

## Tech Stack

- **Backend:** Go + Gin + GORM
- **Database:** PostgreSQL
- **Frontend:** React + Tailwind CSS (planned consumer for metadata-driven UI rendering)
- **Architecture:** Metadata-driven MVC

## Current Features

- Metadata engine for `DocType` and `DocField`
- Automatic PostgreSQL table generation for non-single doctypes
- Seeded system doctypes:
  - `DocType`
  - `DocField`
- Dynamic REST CRUD for records created from metadata
- Field-type-aware payload coercion for:
  - Data/Text
  - Check
  - Int
  - Float/Currency/Percent
  - Date
  - Datetime
  - Time
  - JSON
- Soft delete support via `deleted_at`

## Project Structure

```text
gogal-framework/
├── config/
│   └── db.go
├── controllers/
│   ├── doctype_controller.go
│   └── resource_controller.go
├── models/
│   ├── doctype.go
│   └── resource.go
├── .env.example
├── go.mod
├── main.go
└── README.md
```

## API Overview

### Metadata APIs

- `GET /api/doctypes`
- `POST /api/doctypes`
- `GET /api/doctypes/:name/meta`
- `GET /api/resource-meta/:name`

### Dynamic Resource APIs

- `GET /api/resource/:doctype`
- `POST /api/resource/:doctype`
- `GET /api/resource/:doctype/:name`
- `PUT /api/resource/:doctype/:name`
- `DELETE /api/resource/:doctype/:name`

## Local Setup

### Prerequisites

- Go 1.25+
- PostgreSQL

### Configure environment

Copy the example file and adjust values for your machine:

```bash
cp .env.example .env
```

### Run the API

```bash
go mod tidy
go run main.go
```

The API starts on:

- `http://127.0.0.1:8080`

## Example: Create a DocType

```json
{
  "doctype": "Task",
  "label": "Task",
  "module": "Core",
  "fields": [
    {
      "fieldname": "title",
      "label": "Title",
      "fieldtype": "Data",
      "reqd": true
    },
    {
      "fieldname": "is_done",
      "label": "Is Done",
      "fieldtype": "Check"
    }
  ]
}
```

## Roadmap

- Single DocType storage model
- Dynamic filtering, sorting, and search
- Role-based permissions and access control
- Workflow engine
- Background jobs and scheduler
- React metadata form renderer
- Admin studio for low-code / no-code app building
- Reports, dashboards, and automation

## Status

This repository is in active foundation phase, building toward a full enterprise framework for business app development in Go and JS.

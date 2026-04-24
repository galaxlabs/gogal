# Gogal

Gogal is a Go-based low-code/no-code application platform.

## What Gogal Is
Gogal is a metadata-driven application engine:

Metadata Schema -> Database Migration -> Dynamic CRUD API -> Server-rendered Desk UI -> Workflow/Permissions -> App/SDK Generator -> Cloud + Local Sync.

## Core Vision
- Product name: Gogal
- Platform name: Gogal Platform
- CLI name: `gogal`
- Core Studio runtime: Go Templates + HTMX + Vanilla JS (no React/Vite requirement)

## Quick Start
```bash
cd /home/fg/gogal-framework
go mod tidy
```

## Initialize Project
```bash
go run ./cmd/gogal init .
```

## Create Site
```bash
go run ./cmd/gogal new-site example.local \
  --db-name gogaldb --db-user gogaluser --db-password gogal123 \
  --db-host 127.0.0.1 --db-port 5432
```

## Fix PostgreSQL Permissions
```bash
go run ./cmd/gogal fix-postgres --db gogaldb --user gogaluser
```

## Run Doctor
```bash
go run ./cmd/gogal doctor
```

## Picoclaw
```bash
go run ./cmd/gogal picoclaw diagnose
go run ./cmd/gogal picoclaw fix-postgres --db gogaldb --user gogaluser
```

## Start Server
```bash
go run ./cmd/server
```

## Migration Plan / Apply
```bash
go run ./cmd/gogal migrate --site example.local --plan
go run ./cmd/gogal migrate --site example.local --apply
```

## Create First DocType
Use API:
```bash
curl -X POST http://127.0.0.1:8080/api/doctypes \
  -H 'Content-Type: application/json' \
  -d '{
    "doctype":"Customer",
    "label":"Customer",
    "module":"Core",
    "fields":[
      {"fieldname":"customer_name","fieldtype":"Data","label":"Customer Name","reqd":true}
    ]
  }'
```

## Test APIs
```bash
curl http://127.0.0.1:8080/
curl http://127.0.0.1:8080/api/doctypes
curl http://127.0.0.1:8080/desk
curl http://127.0.0.1:8080/bench
```

## Roadmap Summary
- Phase 0: Foundation + CLI + Postgres privilege tooling
- Phase 1: Metadata core MVP (DocType -> table -> CRUD -> desk)
- Phase 2+: Permissions/workflow, app system, generators, cloud, offline sync, advanced modules

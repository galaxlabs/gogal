# Picoclaw

## What It Is
Picoclaw is Gogal's internal troubleshooting helper.

## How Diagnose Works
Picoclaw runs checks (Go/tooling, config files, PostgreSQL connectivity, schema privileges, create-table capability, port availability) then returns:
- Root cause
- One best fix
- Exact command or SQL
- Verify step
- Next step

## Commands
- `gogal doctor`
- `gogal picoclaw diagnose`
- `gogal picoclaw fix-postgres --db gogaldb --user gogaluser`

## Add New Checks
1. Add a new `CheckResult` entry in `internal/picoclaw/checks.go`.
2. Add diagnosis mapping in `internal/picoclaw/agent.go`.
3. Add remedy helper in `internal/picoclaw/remedies.go`.

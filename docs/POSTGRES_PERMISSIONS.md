# PostgreSQL Permissions (15/16)

## Root Cause
On PostgreSQL 15/16, database access does not guarantee `CREATE` permission on schema `public`.

## Why GORM Migration Fails
GORM can connect but fails on `CREATE TABLE` when the role lacks `USAGE`/`CREATE` on `public`.

## Exact Fix
Use:

```bash
go run ./cmd/gogal fix-postgres --db gogaldb --user gogaluser
```

Manual fallback:

```sql
GRANT CONNECT ON DATABASE gogaldb TO gogaluser;
GRANT ALL PRIVILEGES ON DATABASE gogaldb TO gogaluser;
GRANT USAGE, CREATE ON SCHEMA public TO gogaluser;
ALTER SCHEMA public OWNER TO gogaluser;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gogaluser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gogaluser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO gogaluser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO gogaluser;
```

## Verify
- `go run ./cmd/gogal doctor`
- `go run ./cmd/server`

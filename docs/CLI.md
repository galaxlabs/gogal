# Gogal CLI

## Commands
- `gogal init [bench-name]`
- `gogal new-site [site-name] --db-name --db-user --db-password --db-host --db-port`
- `gogal fix-postgres --db <db> --user <user>`
- `gogal doctor`
- `gogal start`
- `gogal migrate --site [site-name] --plan`
- `gogal migrate --site [site-name] --apply`
- `gogal picoclaw diagnose`
- `gogal picoclaw fix-postgres --db <db> --user <user>`

## Troubleshooting
- If `fix-postgres` cannot run automatically, it prints exact SQL.
- If `doctor` reports schema privilege issues, run `fix-postgres` then rerun `doctor`.
- Use `migrate --plan` before `migrate --apply` to review non-destructive schema actions.

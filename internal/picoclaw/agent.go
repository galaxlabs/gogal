package picoclaw

import "strings"

func DiagnoseFromChecks(results []CheckResult, dbName string, dbUser string) Diagnosis {
	for _, r := range results {
		if r.OK {
			continue
		}
		if strings.Contains(strings.ToLower(r.Name), "port 8080") {
			return Diagnosis{
				Problem: r.Name + ": " + r.Details,
				Fix:     "Run `ss -ltnp | rg :8080` to identify the process, stop it, then restart Gogal.",
				Verify:  "go run ./cmd/gogal doctor",
				Next:    "After port is free, run `go run ./cmd/server`.",
			}
		}
		if strings.Contains(strings.ToLower(r.Name), "schema") || strings.Contains(strings.ToLower(r.Details), "create permission") {
			return PostgresSchemaFix(dbName, dbUser)
		}
		return Diagnosis{
			Problem: r.Name + ": " + r.Details,
			Fix:     "Run `gogal doctor` after fixing this check, or run `gogal picoclaw fix-postgres --db <db> --user <user>` if this is a PostgreSQL privilege issue.",
			Verify:  "go run ./cmd/gogal doctor",
			Next:    "After doctor passes, run `go run ./cmd/server`.",
		}
	}
	return Diagnosis{
		Problem: "No blocking issue detected.",
		Fix:     "No fix required.",
		Verify:  "go run ./cmd/server",
		Next:    "Create a DocType in /desk to continue MVP validation.",
	}
}

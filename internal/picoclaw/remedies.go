package picoclaw

import (
	"fmt"

	"gogal/internal/database"
)

func PostgresSchemaFix(dbName string, dbUser string) Diagnosis {
	return Diagnosis{
		Problem: fmt.Sprintf("GORM cannot create tables because PostgreSQL user %s lacks CREATE permission on schema public.", dbUser),
		Fix:     database.GeneratePostgresGrantSQL(dbName, dbUser),
		Verify:  "go run ./cmd/server",
		Next:    "Run gogal doctor to verify all checks pass.",
	}
}

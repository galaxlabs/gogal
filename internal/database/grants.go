package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func GeneratePostgresGrantSQL(dbName string, dbUser string) string {
	return fmt.Sprintf(`-- Gogal PostgreSQL 15/16 schema privilege fix
GRANT CONNECT ON DATABASE %s TO %s;
GRANT ALL PRIVILEGES ON DATABASE %s TO %s;
GRANT USAGE, CREATE ON SCHEMA public TO %s;
ALTER SCHEMA public OWNER TO %s;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %s;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %s;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s;`,
		quoteIdentifier(dbName), quoteIdentifier(dbUser),
		quoteIdentifier(dbName), quoteIdentifier(dbUser),
		quoteIdentifier(dbUser), quoteIdentifier(dbUser),
		quoteIdentifier(dbUser), quoteIdentifier(dbUser),
		quoteIdentifier(dbUser), quoteIdentifier(dbUser),
	)
}

func CheckSchemaPrivileges(db *sql.DB, dbUser string) error {
	if db == nil {
		return errors.New("nil database handle")
	}
	user := strings.TrimSpace(dbUser)
	if user == "" {
		return errors.New("database user is required")
	}
	var usage bool
	if err := db.QueryRow(`SELECT has_schema_privilege($1, 'public', 'USAGE')`, user).Scan(&usage); err != nil {
		return fmt.Errorf("check USAGE on schema public: %w", err)
	}
	var create bool
	if err := db.QueryRow(`SELECT has_schema_privilege($1, 'public', 'CREATE')`, user).Scan(&create); err != nil {
		return fmt.Errorf("check CREATE on schema public: %w", err)
	}
	if !usage || !create {
		return fmt.Errorf("user %q lacks required schema privileges on public (usage=%t create=%t)", user, usage, create)
	}
	return nil
}

func CanCreateTable(db *sql.DB) error {
	if db == nil {
		return errors.New("nil database handle")
	}
	query := `CREATE TABLE IF NOT EXISTS tab_gogal_perm_check (id BIGSERIAL PRIMARY KEY, created_at TIMESTAMPTZ DEFAULT NOW());
DROP TABLE IF EXISTS tab_gogal_perm_check;`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("cannot create table in public schema: %w", err)
	}
	return nil
}

func quoteIdentifier(v string) string {
	return `"` + strings.ReplaceAll(strings.TrimSpace(v), `"`, `""`) + `"`
}

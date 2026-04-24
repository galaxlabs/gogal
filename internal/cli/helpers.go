package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func writeJSONIfMissing(path string, v any) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return writeJSON(path, v)
}

func writeJSON(path string, v any) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return writeJSONFile(path, v)
}

func writeJSONFile(path string, v any) error {
	b, err := jsonMarshalIndent(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func renderPostgresManualFix(dbName string, dbUser string) string {
	return fmt.Sprintf(`sudo -i -u postgres
psql
\\c %s
GRANT ALL ON SCHEMA public TO %s;
ALTER SCHEMA public OWNER TO %s;
GRANT CREATE, USAGE ON SCHEMA public TO %s;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s;
\\q
exit`, dbName, dbUser, dbUser, dbUser, dbUser, dbUser)
}

func applyPostgresSetup(host string, port int, dbName string, dbUser string, dbPass string) error {
	if _, err := exec.LookPath("psql"); err != nil {
		return err
	}
	superUser := envOrDefault("GOGAL_PG_SUPERUSER", "postgres")
	superPass := os.Getenv("GOGAL_PG_SUPERUSER_PASSWORD")
	roleSQL := fmt.Sprintf(`DO $$ BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '%s') THEN
  CREATE ROLE "%s" LOGIN PASSWORD '%s';
ELSE
  ALTER ROLE "%s" WITH LOGIN PASSWORD '%s';
END IF;
END $$;`, escapeLiteral(dbUser), dbUser, escapeLiteral(dbPass), dbUser, escapeLiteral(dbPass))
	if err := runPSQL(host, port, superUser, superPass, "postgres", roleSQL); err != nil {
		return err
	}
	dbSQL := fmt.Sprintf(`SELECT 'CREATE DATABASE "%s" OWNER "%s"' WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = '%s')\\gexec;
ALTER DATABASE "%s" OWNER TO "%s";`, dbName, dbUser, escapeLiteral(dbName), dbName, dbUser)
	if err := runPSQL(host, port, superUser, superPass, "postgres", dbSQL); err != nil {
		return err
	}
	return applyPostgresGrants(host, port, dbName, dbUser)
}

func applyPostgresGrants(host string, port int, dbName string, dbUser string) error {
	superUser := envOrDefault("GOGAL_PG_SUPERUSER", "postgres")
	superPass := os.Getenv("GOGAL_PG_SUPERUSER_PASSWORD")
	sql := fmt.Sprintf(`GRANT ALL ON SCHEMA public TO "%s";
ALTER SCHEMA public OWNER TO "%s";
GRANT CREATE, USAGE ON SCHEMA public TO "%s";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO "%s";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "%s";`, dbUser, dbUser, dbUser, dbUser, dbUser)
	return runPSQL(host, port, superUser, superPass, dbName, sql)
}

func runPSQL(host string, port int, user string, pass string, db string, sql string) error {
	args := []string{"-v", "ON_ERROR_STOP=1", "-h", host, "-p", fmt.Sprintf("%d", port), "-U", user, "-d", db, "-c", sql}
	cmd := exec.Command("psql", args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+pass)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("psql failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func escapeLiteral(v string) string {
	return strings.ReplaceAll(v, "'", "''")
}

package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const defaultCommandTimeout = 30 * time.Second

var (
	siteNamePattern           = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
	postgresIdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)
)

type siteConfig struct {
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
}

type newSiteOptions struct {
	BenchPath       string
	DBName          string
	DBUser          string
	DBPassword      string
	AdminDBHost     string
	AdminDBPort     int
	AdminDBName     string
	AdminDBUser     string
	AdminDBPassword string
	SkipDBSetup     bool
	NoInput         bool
	CommandTimeout  time.Duration
}

type newSiteResult struct {
	BenchRoot      string
	SiteName       string
	SitePath       string
	SiteConfigPath string
	DBName         string
	DBUser         string
	DBSetupSkipped bool
}

func newNewSiteCommand() *cobra.Command {
	options := &newSiteOptions{}

	cmd := &cobra.Command{
		Use:   "new-site [site-name]",
		Short: "Create a new site inside a Gogal bench",
		Long: strings.TrimSpace(`Create a new multi-tenant site with site-specific folders and site_config.json.

By default the command also provisions a PostgreSQL database and role using psql. It is safe to run multiple times: existing directories are reused, existing site config is preserved, and database creation statements are idempotent.`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := createSite(cmd, args[0], options)
			if err != nil {
				return err
			}

			cmd.Printf("Created site %s in bench %s\n", result.SiteName, result.BenchRoot)
			cmd.Printf("Site path: %s\n", result.SitePath)
			cmd.Printf("Site config: %s\n", result.SiteConfigPath)
			cmd.Printf("Database: %s\n", result.DBName)
			cmd.Printf("Database user: %s\n", result.DBUser)
			if result.DBSetupSkipped {
				cmd.Println("Database provisioning: skipped")
			} else {
				cmd.Println("Database provisioning: completed")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&options.BenchPath, "bench", ".", "Bench root directory")
	cmd.Flags().StringVar(&options.DBName, "db-name", "", "Site database name (defaults to an auto-generated value)")
	cmd.Flags().StringVar(&options.DBUser, "db-user", "", "Site database user (defaults to an auto-generated value)")
	cmd.Flags().StringVar(&options.DBPassword, "db-password", "", "Site database password (defaults to an auto-generated value)")
	cmd.Flags().StringVar(&options.AdminDBHost, "admin-db-host", "", "Administrative PostgreSQL host (defaults to common_site_config db_host)")
	cmd.Flags().IntVar(&options.AdminDBPort, "admin-db-port", 0, "Administrative PostgreSQL port (defaults to common_site_config db_port)")
	cmd.Flags().StringVar(&options.AdminDBName, "admin-db-name", "postgres", "Administrative PostgreSQL database")
	cmd.Flags().StringVar(&options.AdminDBUser, "admin-db-user", envOrDefault("GOGAL_PG_SUPERUSER", envOrDefault("POSTGRES_USER", "postgres")), "Administrative PostgreSQL user")
	cmd.Flags().StringVar(&options.AdminDBPassword, "admin-db-password", envOrDefault("GOGAL_PG_SUPERUSER_PASSWORD", os.Getenv("PGPASSWORD")), "Administrative PostgreSQL password")
	cmd.Flags().BoolVar(&options.SkipDBSetup, "skip-db-setup", false, "Skip PostgreSQL role/database creation and only scaffold site files")
	cmd.Flags().BoolVar(&options.NoInput, "no-input", false, "Disable interactive prompts and auto-generate missing values")
	cmd.Flags().DurationVar(&options.CommandTimeout, "command-timeout", defaultCommandTimeout, "Timeout for PostgreSQL shell commands")

	return cmd
}

func createSite(cmd *cobra.Command, rawSiteName string, options *newSiteOptions) (*newSiteResult, error) {
	siteName := strings.TrimSpace(rawSiteName)
	if !siteNamePattern.MatchString(siteName) {
		return nil, fmt.Errorf("invalid site name %q: use letters, numbers, dots, dashes, or underscores", rawSiteName)
	}

	benchRoot, err := filepath.Abs(filepath.Clean(options.BenchPath))
	if err != nil {
		return nil, fmt.Errorf("resolve bench path: %w", err)
	}

	if _, err := initializeBench(benchRoot); err != nil {
		return nil, err
	}

	commonConfigPath := filepath.Join(benchRoot, "sites", "common_site_config.json")
	commonConfig, err := readCommonSiteConfig(commonConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read common site config: %w", err)
	}

	sitePath := filepath.Join(benchRoot, "sites", siteName)
	publicPath := filepath.Join(sitePath, "public")
	privatePath := filepath.Join(sitePath, "private")
	for _, dir := range []string{sitePath, publicPath, privatePath} {
		if err := ensureDirectory(dir); err != nil {
			return nil, err
		}
	}

	siteConfigPath := filepath.Join(sitePath, "site_config.json")
	existingSiteConfig, err := readSiteConfig(siteConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	identifierBase := makePostgresIdentifierBase(siteName)

	resolvedConfig := &siteConfig{
		DBHost: commonConfig.DBHost,
		DBPort: commonConfig.DBPort,
	}
	if existingSiteConfig != nil {
		mergeSiteConfig(resolvedConfig, existingSiteConfig)
	}

	resolvedConfig.DBName = firstNonEmpty(options.DBName, resolvedConfig.DBName)
	if resolvedConfig.DBName == "" {
		resolvedConfig.DBName = truncatePostgresIdentifier(identifierBase)
	}
	if resolvedConfig.DBName, err = promptValue(cmd, reader, "Database name", resolvedConfig.DBName, options.NoInput); err != nil {
		return nil, err
	}
	if err := validatePostgresIdentifier(resolvedConfig.DBName, "database name"); err != nil {
		return nil, err
	}

	resolvedConfig.DBUser = firstNonEmpty(options.DBUser, resolvedConfig.DBUser)
	if resolvedConfig.DBUser == "" {
		resolvedConfig.DBUser = truncatePostgresIdentifier(identifierBase + "_user")
	}
	if resolvedConfig.DBUser, err = promptValue(cmd, reader, "Database user", resolvedConfig.DBUser, options.NoInput); err != nil {
		return nil, err
	}
	if err := validatePostgresIdentifier(resolvedConfig.DBUser, "database user"); err != nil {
		return nil, err
	}

	resolvedConfig.DBPassword = firstNonEmpty(options.DBPassword, resolvedConfig.DBPassword)
	if resolvedConfig.DBPassword == "" {
		resolvedConfig.DBPassword, err = generatePassword(24)
		if err != nil {
			return nil, err
		}
	}
	if resolvedConfig.DBPassword, err = promptValue(cmd, reader, "Database password", resolvedConfig.DBPassword, options.NoInput); err != nil {
		return nil, err
	}

	resolvedConfig.DBHost = firstNonEmpty(resolvedConfig.DBHost, commonConfig.DBHost)
	if options.AdminDBHost == "" {
		options.AdminDBHost = resolvedConfig.DBHost
	}
	if options.AdminDBPort == 0 {
		options.AdminDBPort = resolvedConfig.DBPort
	}

	if err := writeJSONFile(siteConfigPath, resolvedConfig); err != nil {
		return nil, err
	}

	if !options.SkipDBSetup {
		postgresOptions := postgresProvisionOptions{
			Host:        options.AdminDBHost,
			Port:        options.AdminDBPort,
			Database:    options.AdminDBName,
			AdminUser:   options.AdminDBUser,
			AdminPass:   options.AdminDBPassword,
			CommandWait: options.CommandTimeout,
		}

		if postgresOptions.AdminUser == "" {
			postgresOptions.AdminUser = "postgres"
		}
		if postgresOptions.Database == "" {
			postgresOptions.Database = "postgres"
		}
		if postgresOptions.Host == "" {
			postgresOptions.Host = commonConfig.DBHost
		}
		if postgresOptions.Port == 0 {
			postgresOptions.Port = commonConfig.DBPort
		}

		if err := provisionPostgresSite(postgresOptions, resolvedConfig); err != nil {
			return nil, err
		}
	}

	return &newSiteResult{
		BenchRoot:      benchRoot,
		SiteName:       siteName,
		SitePath:       sitePath,
		SiteConfigPath: siteConfigPath,
		DBName:         resolvedConfig.DBName,
		DBUser:         resolvedConfig.DBUser,
		DBSetupSkipped: options.SkipDBSetup,
	}, nil
}

type postgresProvisionOptions struct {
	Host        string
	Port        int
	Database    string
	AdminUser   string
	AdminPass   string
	CommandWait time.Duration
}

func provisionPostgresSite(options postgresProvisionOptions, site *siteConfig) error {
	if _, err := exec.LookPath("psql"); err != nil {
		return fmt.Errorf("psql is required for site database provisioning: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.CommandWait)
	defer cancel()

	roleSQL := fmt.Sprintf(`DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = %s) THEN
    CREATE ROLE %s LOGIN PASSWORD %s;
  ELSE
    ALTER ROLE %s WITH LOGIN PASSWORD %s;
  END IF;
END
$$;`, quoteLiteral(site.DBUser), quoteIdentifier(site.DBUser), quoteLiteral(site.DBPassword), quoteIdentifier(site.DBUser), quoteLiteral(site.DBPassword))

	if err := runPSQLCommand(ctx, options, roleSQL, options.Database); err != nil {
		return fmt.Errorf("ensure database role %s: %w", site.DBUser, err)
	}

	databaseSQL := fmt.Sprintf(`SELECT 'CREATE DATABASE %s OWNER %s'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = %s)\gexec
ALTER DATABASE %s OWNER TO %s;
GRANT ALL PRIVILEGES ON DATABASE %s TO %s;`, quoteIdentifier(site.DBName), quoteIdentifier(site.DBUser), quoteLiteral(site.DBName), quoteIdentifier(site.DBName), quoteIdentifier(site.DBUser), quoteIdentifier(site.DBName), quoteIdentifier(site.DBUser))

	if err := runPSQLCommand(ctx, options, databaseSQL, options.Database); err != nil {
		return fmt.Errorf("ensure database %s: %w", site.DBName, err)
	}

	schemaSQL := fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s;
GRANT ALL ON SCHEMA public TO %s;
ALTER SCHEMA public OWNER TO %s;`, quoteIdentifier(site.DBName), quoteIdentifier(site.DBUser), quoteIdentifier(site.DBUser), quoteIdentifier(site.DBUser))

	if err := runPSQLCommand(ctx, options, schemaSQL, site.DBName); err != nil {
		return fmt.Errorf("grant schema privileges for %s: %w", site.DBUser, err)
	}

	return nil
}

func runPSQLCommand(ctx context.Context, options postgresProvisionOptions, sql string, database string) error {
	args := []string{
		"-v", "ON_ERROR_STOP=1",
		"-h", options.Host,
		"-p", fmt.Sprintf("%d", options.Port),
		"-U", options.AdminUser,
		"-d", database,
		"-c", sql,
	}

	command := exec.CommandContext(ctx, "psql", args...)
	command.Env = append(os.Environ(), "PGPASSWORD="+options.AdminPass)
	output, err := command.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("psql command timed out after %s", options.CommandWait)
		}
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func readSiteConfig(path string) (*siteConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config siteConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &config, nil
}

func mergeSiteConfig(target, existing *siteConfig) {
	if strings.TrimSpace(existing.DBName) != "" {
		target.DBName = existing.DBName
	}
	if strings.TrimSpace(existing.DBUser) != "" {
		target.DBUser = existing.DBUser
	}
	if strings.TrimSpace(existing.DBPassword) != "" {
		target.DBPassword = existing.DBPassword
	}
	if strings.TrimSpace(existing.DBHost) != "" {
		target.DBHost = existing.DBHost
	}
	if existing.DBPort != 0 {
		target.DBPort = existing.DBPort
	}
}

func promptValue(cmd *cobra.Command, reader *bufio.Reader, label, currentValue string, noInput bool) (string, error) {
	if noInput || !isInteractiveTerminal() {
		return currentValue, nil
	}

	defaultSuffix := ""
	if currentValue != "" {
		defaultSuffix = fmt.Sprintf(" [%s]", currentValue)
	}
	cmd.Printf("%s%s: ", label, defaultSuffix)

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, os.ErrClosed) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			return currentValue, nil
		}
		return trimmed, nil
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return currentValue, nil
	}

	return trimmed, nil
}

func isInteractiveTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func makePostgresIdentifierBase(siteName string) string {
	normalized := strings.ToLower(strings.TrimSpace(siteName))
	normalized = strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(normalized)
	normalized = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(normalized, "")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		normalized = "site"
	}
	if normalized[0] >= '0' && normalized[0] <= '9' {
		normalized = "site_" + normalized
	}
	return normalized
}

func truncatePostgresIdentifier(value string) string {
	if len(value) <= 63 {
		return value
	}
	return strings.TrimRight(value[:63], "_")
}

func validatePostgresIdentifier(value string, label string) error {
	if !postgresIdentifierPattern.MatchString(value) {
		return fmt.Errorf("invalid %s %q: use letters, numbers, and underscores; maximum length is 63 characters", label, value)
	}
	return nil
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func quoteLiteral(value string) string {
	return `'` + strings.ReplaceAll(value, `'`, `''`) + `'`
}

func generatePassword(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

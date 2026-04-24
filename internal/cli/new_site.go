package cli

import (
	"fmt"
	"path/filepath"

	"gogal/internal/config"

	"github.com/spf13/cobra"
)

func newNewSiteCommand() *cobra.Command {
	var (
		root       string
		dbName     string
		dbUser     string
		dbPassword string
		dbHost     string
		dbPort     int
		createDB   bool
	)
	cmd := &cobra.Command{
		Use:   "new-site [site-name]",
		Short: "Create a new Gogal site",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if root == "" {
				root = "."
			}
			absRoot, err := filepath.Abs(filepath.Clean(root))
			if err != nil {
				return err
			}
			siteName := args[0]
			siteCfg := config.DefaultSiteConfig(siteName)
			if dbName != "" {
				siteCfg.DBName = dbName
			}
			if dbUser != "" {
				siteCfg.DBUser = dbUser
			}
			if dbPassword != "" {
				siteCfg.DBPassword = dbPassword
			}
			if dbHost != "" {
				siteCfg.DBHost = dbHost
			}
			if dbPort != 0 {
				siteCfg.DBPort = dbPort
			}
			if err := ensureDir(filepath.Join(absRoot, "sites", siteName)); err != nil {
				return err
			}
			if err := writeJSON(filepath.Join(absRoot, "sites", siteName, "site_config.json"), siteCfg); err != nil {
				return err
			}
			if createDB {
				if err := applyPostgresSetup(siteCfg.DBHost, siteCfg.DBPort, siteCfg.DBName, siteCfg.DBUser, siteCfg.DBPassword); err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Database auto-setup failed: %v\n", err)
					fmt.Fprintln(cmd.OutOrStdout(), "Run manually:")
					fmt.Fprintln(cmd.OutOrStdout(), renderPostgresManualFix(siteCfg.DBName, siteCfg.DBUser))
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Site created: %s\n", filepath.Join(absRoot, "sites", siteName, "site_config.json"))
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "Bench/project root")
	cmd.Flags().StringVar(&dbName, "db-name", "", "Database name")
	cmd.Flags().StringVar(&dbUser, "db-user", "", "Database user")
	cmd.Flags().StringVar(&dbPassword, "db-password", "", "Database password")
	cmd.Flags().StringVar(&dbHost, "db-host", "", "Database host")
	cmd.Flags().IntVar(&dbPort, "db-port", 0, "Database port")
	cmd.Flags().BoolVar(&createDB, "create-db", true, "Create database/user and apply grants when possible")
	return cmd
}

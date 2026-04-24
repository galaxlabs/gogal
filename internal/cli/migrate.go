package cli

import (
	"fmt"
	"os"

	"gogal/internal/config"
	"gogal/internal/database"
	"gogal/internal/migration"

	"github.com/spf13/cobra"
)

func newMigrateCommand() *cobra.Command {
	var site string
	var planOnly bool
	var apply bool
	cmd := &cobra.Command{
		Use:   "migrate --site [site-name] --plan|--apply",
		Short: "Plan or apply metadata-driven migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if site == "" {
				site = "example.local"
			}
			rc, err := config.LoadRuntimeConfigFromRoot(root)
			if err != nil {
				return err
			}
			rc.SiteName = site
			_ = os.Setenv("DB_HOST", rc.Site.DBHost)
			_ = os.Setenv("DB_PORT", fmt.Sprintf("%d", rc.Site.DBPort))
			_ = os.Setenv("DB_NAME", rc.Site.DBName)
			_ = os.Setenv("DB_USER", rc.Site.DBUser)
			_ = os.Setenv("DB_PASSWORD", rc.Site.DBPassword)
			dbCfg := database.DBConfig{
				Host:     rc.Site.DBHost,
				Port:     rc.Site.DBPort,
				Name:     rc.Site.DBName,
				User:     rc.Site.DBUser,
				Password: rc.Site.DBPassword,
			}
			if _, err := database.Connect(dbCfg); err != nil {
				return err
			}

			exec := migration.NewExecutor(database.DB)
			plan, err := exec.BuildPlan()
			if err != nil {
				return err
			}

			if !planOnly && !apply {
				planOnly = true
			}
			if planOnly && apply {
				return fmt.Errorf("use only one mode: --plan or --apply")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Migration Plan")
			if len(plan.Items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "- No schema changes required.")
			} else {
				for _, item := range plan.Items {
					switch item.Type {
					case migration.PlanCreateTable:
						fmt.Fprintf(cmd.OutOrStdout(), "- table to create: %s (DocType: %s)\n", item.Table, item.DocType)
					case migration.PlanAddColumn:
						fmt.Fprintf(cmd.OutOrStdout(), "- column to add: %s.%s (DocType: %s)\n", item.Table, item.Column, item.DocType)
					case migration.PlanAddUnique:
						fmt.Fprintf(cmd.OutOrStdout(), "- unique to add: %s.%s (DocType: %s)\n", item.Table, item.Column, item.DocType)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  SQL: %s\n", item.Statement)
				}
			}
			for _, warning := range plan.UnsafeWarnings {
				fmt.Fprintf(cmd.OutOrStdout(), "WARNING: %s\n", warning)
			}

			if apply {
				if err := exec.ApplyPlan(plan); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Migration apply completed.")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Plan only mode completed. Re-run with --apply to execute.")
			return nil
		},
	}
	cmd.Flags().StringVar(&site, "site", "example.local", "Site name")
	cmd.Flags().BoolVar(&planOnly, "plan", false, "Preview migration changes without applying")
	cmd.Flags().BoolVar(&apply, "apply", false, "Apply migration changes")
	return cmd
}

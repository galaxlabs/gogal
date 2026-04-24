package cli

import (
	"fmt"

	"gogal/internal/database"

	"github.com/spf13/cobra"
)

func newFixPostgresCommand() *cobra.Command {
	var dbName string
	var dbUser string
	var host string
	var port int
	cmd := &cobra.Command{
		Use:   "fix-postgres --db <db> --user <user>",
		Short: "Fix PostgreSQL schema privileges for Gogal",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbName == "" || dbUser == "" {
				return fmt.Errorf("--db and --user are required")
			}
			if err := applyPostgresGrants(host, port, dbName, dbUser); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Automatic grant execution failed: %v\n\n", err)
				fmt.Fprintln(cmd.OutOrStdout(), "Run this SQL manually as postgres superuser:")
				fmt.Fprintln(cmd.OutOrStdout(), database.GeneratePostgresGrantSQL(dbName, dbUser))
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "PostgreSQL schema grants applied successfully.")
			return nil
		},
	}
	cmd.Flags().StringVar(&dbName, "db", "", "Database name")
	cmd.Flags().StringVar(&dbUser, "user", "", "Database user")
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "PostgreSQL host")
	cmd.Flags().IntVar(&port, "port", 5432, "PostgreSQL port")
	_ = cmd.MarkFlagRequired("db")
	_ = cmd.MarkFlagRequired("user")
	return cmd
}

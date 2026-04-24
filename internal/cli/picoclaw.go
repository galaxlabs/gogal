package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"gogal/internal/picoclaw"

	"github.com/spf13/cobra"
)

func newPicoclawCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "picoclaw", Short: "Picoclaw helper commands"}

	diagnose := &cobra.Command{
		Use:   "diagnose",
		Short: "Run Picoclaw diagnosis",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			root, _ := filepath.Abs(wd)
			results, err := picoclaw.RunDoctorChecks(root, "example.local")
			if err != nil {
				return err
			}
			d := picoclaw.DiagnoseFromChecks(results, "gogaldb", "gogaluser")
			fmt.Fprintf(cmd.OutOrStdout(), "Problem:\n%s\n\nFix:\n%s\n\nVerify:\n%s\n\nNext:\n%s\n", d.Problem, d.Fix, d.Verify, d.Next)
			return nil
		},
	}

	var dbName, dbUser string
	fix := &cobra.Command{
		Use:   "fix-postgres",
		Short: "Print Picoclaw PostgreSQL fix",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := picoclaw.PostgresSchemaFix(dbName, dbUser)
			fmt.Fprintf(cmd.OutOrStdout(), "Problem:\n%s\n\nFix:\n%s\n\nVerify:\n%s\n\nNext:\n%s\n", d.Problem, d.Fix, d.Verify, d.Next)
			return nil
		},
	}
	fix.Flags().StringVar(&dbName, "db", "gogaldb", "Database name")
	fix.Flags().StringVar(&dbUser, "user", "gogaluser", "Database user")

	cmd.AddCommand(diagnose, fix)
	return cmd
}

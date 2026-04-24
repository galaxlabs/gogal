package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gogal/internal/picoclaw"

	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	var site string
	var root string
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose Gogal environment issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			if root == "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				root = wd
			}
			absRoot, err := filepath.Abs(filepath.Clean(root))
			if err != nil {
				return err
			}
			if site == "" {
				site = os.Getenv("GOGAL_SITE")
				if site == "" {
					site = "example.local"
				}
			}
			results, err := picoclaw.RunDoctorChecks(absRoot, site)
			if err != nil {
				return err
			}
			okAll := true
			for _, r := range results {
				status := "OK"
				if !r.OK {
					status = "FAIL"
					okAll = false
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", status, r.Name, r.Details)
			}
			diag := picoclaw.DiagnoseFromChecks(results, "gogaldb", "gogaluser")
			fmt.Fprintln(cmd.OutOrStdout(), "\nPicoclaw Diagnosis")
			fmt.Fprintf(cmd.OutOrStdout(), "Problem: %s\n", diag.Problem)
			fmt.Fprintf(cmd.OutOrStdout(), "Fix: %s\n", diag.Fix)
			fmt.Fprintf(cmd.OutOrStdout(), "Verify: %s\n", diag.Verify)
			fmt.Fprintf(cmd.OutOrStdout(), "Next: %s\n", diag.Next)
			if okAll {
				fmt.Fprintln(cmd.OutOrStdout(), "\nNext command: go run ./cmd/server")
			} else {
				if strings.Contains(strings.ToLower(diag.Problem), "postgres") || strings.Contains(strings.ToLower(diag.Problem), "schema") {
					fmt.Fprintln(cmd.OutOrStdout(), "\nNext command: go run ./cmd/gogal fix-postgres --db gogaldb --user gogaluser")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "\nNext command: go run ./cmd/gogal doctor")
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "Project root")
	cmd.Flags().StringVar(&site, "site", "", "Site name")
	return cmd
}

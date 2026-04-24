package cli

import (
	"fmt"
	"path/filepath"

	"gogal/internal/config"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [bench-name]",
		Short: "Initialize Gogal bench folders",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := filepath.Abs(filepath.Clean(args[0]))
			if err != nil {
				return err
			}
			dirs := []string{
				"apps", "sites", "config", "public", "private", "views", "scripts", "docs",
				"apps/core/doctypes", "sites/example.local", "public/css", "public/js", "views/partials",
			}
			for _, dir := range dirs {
				if err := ensureDir(filepath.Join(root, dir)); err != nil {
					return err
				}
			}
			commonPath := filepath.Join(root, "sites", "common_site_config.json")
			if err := writeJSONIfMissing(commonPath, config.DefaultCommonSiteConfig()); err != nil {
				return err
			}
			sitePath := filepath.Join(root, "sites", "example.local", "site_config.json")
			if err := writeJSONIfMissing(sitePath, config.DefaultSiteConfig("example.local")); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Gogal bench initialized at %s\n", root)
			return nil
		},
	}
}

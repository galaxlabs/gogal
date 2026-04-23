package main

import "github.com/spf13/cobra"

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "gogal",
		Short:         "Gogal bench and site management CLI",
		Long:          "gogal-cli manages benches, apps, sites, development workflows, and production setup for Gogal.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newInstallAppCommand())
	rootCmd.AddCommand(newNewAppCommand())
	rootCmd.AddCommand(newNewSiteCommand())

	return rootCmd
}

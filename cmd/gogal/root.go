package main

import "github.com/spf13/cobra"

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "gogal",
		Short:         "Gogal Framework bench and site management CLI",
		Long:          "gogal-cli manages benches, apps, sites, development workflows, and production setup for Gogal Framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newInitCommand())

	return rootCmd
}

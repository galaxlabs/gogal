package cli

import "github.com/spf13/cobra"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gogal",
		Short:         "Gogal Platform CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newNewSiteCommand())
	cmd.AddCommand(newFixPostgresCommand())
	cmd.AddCommand(newDoctorCommand())
	cmd.AddCommand(newStartCommand())
	cmd.AddCommand(newMigrateCommand())
	cmd.AddCommand(newPicoclawCommand())
	return cmd
}

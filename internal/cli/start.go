package cli

import (
	"context"

	"gogal/internal/app"

	"github.com/spf13/cobra"
)

func newStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start Gogal server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.RunServer(context.Background())
		},
	}
}

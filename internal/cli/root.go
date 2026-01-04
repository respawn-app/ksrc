package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ksrc",
		Short:        "Kotlin dependency source search",
		SilenceUsage: true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newSearchCmd(app))
	cmd.AddCommand(newCatCmd(app))
	cmd.AddCommand(newOpenCmd(app))
	cmd.AddCommand(newDepsCmd(app))
	cmd.AddCommand(newResolveCmd(app))
	cmd.AddCommand(newFetchCmd(app))
	cmd.AddCommand(newWhereCmd(app))
	cmd.AddCommand(newDoctorCmd(app))

	return cmd
}

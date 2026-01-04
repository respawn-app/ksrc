package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/respawn-app/ksrc/internal/search"
	"github.com/spf13/cobra"
)

func newSearchCmd(app *App) *cobra.Command {
	var flags ResolveFlags
	var query string
	var rgArgs string

	cmd := &cobra.Command{
		Use:     "search [<module>]",
		Aliases: []string{"rg"},
		Short:   "Search dependency sources",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				if flags.Module != "" && flags.Module != args[0] {
					return fmt.Errorf("module specified twice (arg and --module). Use only one.")
				}
				flags.Module = args[0]
			}
			if err := requireModuleOrAll(flags.Module, flags.All); err != nil {
				return err
			}
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("query is required. Try: ksrc search --all -q \"<pattern>\"")
			}
			ctx := context.Background()
			sources, _, err := resolveSources(ctx, app, flags, "", true, true)
			if err != nil {
				return err
			}
			if len(sources) == 0 {
				return noSourcesErr(flags, noSourcesHintForFlags(flags))
			}
			matches, err := search.Run(ctx, app.Runner, search.Options{
				Pattern: query,
				Jars:    sources,
				RGArgs:  splitCSV(rgArgs),
				WorkDir: flags.Project,
			})
			if err != nil {
				return err
			}
			for _, m := range matches {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s:%d:%d:%s\n", m.FileID, m.File, m.Line, m.Column, m.Text)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "search query (required)")
	cmd.Flags().StringVar(&flags.Project, "project", ".", "project root")
	cmd.Flags().BoolVar(&flags.All, "all", false, "search all resolved dependencies")
	cmd.Flags().StringVar(&flags.Module, "module", "", "module selector (group:artifact[:version])")
	cmd.Flags().StringVar(&flags.Group, "group", "", "group filter")
	cmd.Flags().StringVar(&flags.Artifact, "artifact", "", "artifact filter")
	cmd.Flags().StringVar(&flags.Version, "version", "", "version filter")
	cmd.Flags().StringVar(&flags.Scope, "scope", "compile", "dependency scope (compile|runtime|test|all)")
	cmd.Flags().StringVar(&flags.Config, "config", "", "configuration name(s) (comma-separated)")
	cmd.Flags().StringVar(&flags.Targets, "targets", "", "KMP targets (comma-separated)")
	cmd.Flags().StringSliceVar(&flags.Subprojects, "subproject", nil, "limit to subproject (repeatable)")
	cmd.Flags().BoolVar(&flags.Offline, "offline", false, "offline mode")
	cmd.Flags().BoolVar(&flags.Refresh, "refresh", false, "refresh dependencies")
	cmd.Flags().StringVar(&rgArgs, "rg-args", "", "extra args for rg (comma-separated)")

	return cmd
}

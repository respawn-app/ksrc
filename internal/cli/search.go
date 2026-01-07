package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/respawn-app/ksrc/internal/search"
	"github.com/spf13/cobra"
)

func newSearchCmd(app *App) *cobra.Command {
	var flags ResolveFlags
	var query string
	var rgArgs string
	var showExtractedPath bool
	var contextLines int

	cmd := &cobra.Command{
		Use:     "search [<module>] [-- <rg-args>]",
		Aliases: []string{"rg"},
		Short:   "Search dependency sources",
		Args: func(cmd *cobra.Command, args []string) error {
			dash := cmd.Flags().ArgsLenAtDash()
			if dash == -1 {
				return cobra.MaximumNArgs(1)(cmd, args)
			}
			if dash > 1 {
				return fmt.Errorf("expected at most one <module> before --")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dash := cmd.Flags().ArgsLenAtDash()
			var moduleArg string
			var passArgs []string
			if dash >= 0 {
				if dash > len(args) {
					dash = len(args)
				}
				if dash == 1 {
					moduleArg = args[0]
				}
				passArgs = append(passArgs, args[dash:]...)
			} else if len(args) == 1 {
				moduleArg = args[0]
			}
			if moduleArg != "" {
				if flags.Module != "" && flags.Module != moduleArg {
					return fmt.Errorf("module specified twice (arg and --module). Use only one.")
				}
				flags.Module = moduleArg
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
			rgExtra := splitCSV(rgArgs)
			if contextLines > 0 {
				rgExtra = append(rgExtra, "-C", strconv.Itoa(contextLines))
			}
			rgExtra = append(rgExtra, passArgs...)
			matches, err := search.Run(ctx, app.Runner, search.Options{
				Pattern: query,
				Jars:    sources,
				RGArgs:  rgExtra,
				WorkDir: flags.Project,
			})
			if err != nil {
				return err
			}
			for _, m := range matches {
				if showExtractedPath {
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s:%d:%d:%s\n", m.FileID, m.File, m.Line, m.Column, m.Text)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %d:%d:%s\n", m.FileID, m.Line, m.Column, m.Text)
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
	cmd.Flags().BoolVar(&flags.IncludeBuildSrc, "buildsrc", true, "include buildSrc dependencies (set --buildsrc=false to disable)")
	cmd.Flags().StringVar(&rgArgs, "rg-args", "", "extra args for rg (comma-separated)")
	cmd.Flags().BoolVar(&showExtractedPath, "show-extracted-path", false, "include temp extracted path in output")
	cmd.Flags().IntVar(&contextLines, "context", 0, "show N lines before/after matches (rg -C)")

	return cmd
}

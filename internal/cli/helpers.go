package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/respawn-app/ksrc/internal/gradle"
	"github.com/respawn-app/ksrc/internal/resolve"
)

func resolveSources(ctx context.Context, app *App, flags ResolveFlags, dep string, applyFilters bool, allowCacheFallback bool) ([]resolve.SourceJar, []resolve.Coord, error) {
	if strings.TrimSpace(flags.Project) == "" {
		flags.Project = "."
	}
	opts := flags.ToOptions()
	opts.Dep = dep
	res, err := gradle.Resolve(ctx, app.Runner, opts)
	if err != nil {
		return nil, nil, err
	}
	sources := res.Sources
	if applyFilters {
		sources = resolve.FilterSources(sources, flags.Module, flags.Group, flags.Artifact, flags.Version)
	}
	if len(sources) == 0 && allowCacheFallback {
		if coord, ok := resolve.SelectorToCoord(flags.Module, flags.Group, flags.Artifact, flags.Version); ok {
			if coord.Version == "" {
				cached, err := resolve.FindCachedSources(coord.Group, coord.Artifact, "")
				if err == nil {
					sources = cached
				}
			}
		}
	}
	return sources, res.Deps, nil
}

func requireModuleOrAll(module string, all bool) error {
	if strings.TrimSpace(module) == "" && !all {
		return fmt.Errorf("E_NO_MODULE: <module> required unless --all is provided")
	}
	return nil
}

package cli

import (
	"context"
	"errors"
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
		return fmt.Errorf("E_NO_MODULE: <module> required unless --all is provided. Try: ksrc search --all -q \"<pattern>\" or ksrc search group:artifact -q \"<pattern>\"")
	}
	return nil
}

func noSourcesErr(flags ResolveFlags, hint string) error {
	msg := "E_NO_SOURCES: no sources resolved."
	var parts []string
	if flags.Offline {
		parts = append(parts, "You ran with --offline; rerun without it to allow downloads.")
	}
	if strings.TrimSpace(hint) != "" {
		parts = append(parts, hint)
	}
	if len(parts) == 0 {
		return errors.New(strings.TrimSuffix(msg, "."))
	}
	return errors.New(msg + " " + strings.Join(parts, " "))
}

func noSourcesHintForFlags(flags ResolveFlags) string {
	if flags.All {
		return "Try: ksrc deps (list resolved coords), then ksrc fetch <coord> to download sources."
	}
	if coord, ok := resolve.SelectorToCoord(flags.Module, flags.Group, flags.Artifact, flags.Version); ok {
		if coord.Version != "" {
			return fmt.Sprintf("Try: ksrc fetch %s to download sources.", coord.String())
		}
		if coord.Group != "" || coord.Artifact != "" {
			return "Try: add a version (group:artifact:version) or run ksrc deps to see resolved coords."
		}
	}
	return "Try: ksrc deps (list resolved coords), then ksrc fetch <coord> to download sources."
}

func noSourcesHintForCoord(coord resolve.Coord) string {
	if coord.Version != "" {
		return fmt.Sprintf("Try: ksrc fetch %s to download sources.", coord.String())
	}
	return "Try: ksrc deps (list resolved coords), then ksrc fetch <coord> to download sources."
}

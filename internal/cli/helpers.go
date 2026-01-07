package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/respawn-app/ksrc/internal/gradle"
	"github.com/respawn-app/ksrc/internal/resolve"
)

type ResolveMeta struct {
	Attempts            []string
	TriedConfigPatterns []string
}

func resolveSources(ctx context.Context, app *App, flags ResolveFlags, dep string, applyFilters bool, allowCacheFallback bool) ([]resolve.SourceJar, []resolve.Coord, ResolveMeta, error) {
	if strings.TrimSpace(flags.Project) == "" {
		flags.Project = "."
	}
	opts := flags.ToOptions()
	opts.Dep = dep

	meta := ResolveMeta{}
	attempts := buildResolveAttempts(opts, flags)
	var lastDeps []resolve.Coord
	for _, attempt := range attempts {
		res, err := gradle.Resolve(ctx, app.Runner, attempt.Options)
		if err != nil {
			return nil, nil, meta, err
		}
		meta.Attempts = append(meta.Attempts, attempt.Label)
		meta.TriedConfigPatterns = append(meta.TriedConfigPatterns, attempt.ConfigPatterns...)
		lastDeps = res.Deps
		sources := res.Sources
		if applyFilters {
			sources = resolve.FilterSources(sources, flags.Module, flags.Group, flags.Artifact, flags.Version)
		}
		if len(sources) > 0 || (!applyFilters && len(res.Deps) > 0) {
			return sources, res.Deps, meta, nil
		}
	}

	var sources []resolve.SourceJar
	if allowCacheFallback {
		if coord, ok := resolve.SelectorToCoord(flags.Module, flags.Group, flags.Artifact, flags.Version); ok {
			if coord.Version == "" {
				cached, err := resolve.FindCachedSources(coord.Group, coord.Artifact, "")
				if err == nil {
					sources = cached
				}
			}
		}
	}
	return sources, lastDeps, meta, nil
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

func noSourcesHintForFlags(flags ResolveFlags, meta ResolveMeta) string {
	if flags.All {
		return joinHints(
			"Try: ksrc deps (list resolved coords), then ksrc fetch <coord> to download sources.",
			projectHint(flags, meta),
		)
	}
	if coord, ok := resolve.SelectorToCoord(flags.Module, flags.Group, flags.Artifact, flags.Version); ok {
		if coord.Version != "" {
			return joinHints(
				fmt.Sprintf("Try: ksrc fetch %s to download sources.", coord.String()),
				projectHint(flags, meta),
			)
		}
		if coord.Group != "" || coord.Artifact != "" {
			return joinHints(
				"Try: add a version (group:artifact:version) or run ksrc deps to see resolved coords.",
				projectHint(flags, meta),
			)
		}
	}
	return joinHints(
		"Try: ksrc deps (list resolved coords), then ksrc fetch <coord> to download sources.",
		projectHint(flags, meta),
	)
}

func noSourcesHintForCoord(coord resolve.Coord) string {
	if coord.Version != "" {
		return fmt.Sprintf("Try: ksrc fetch %s to download sources.", coord.String())
	}
	return "Try: ksrc deps (list resolved coords), then ksrc fetch <coord> to download sources."
}

type resolveAttempt struct {
	Options        gradle.ResolveOptions
	Label          string
	ConfigPatterns []string
}

func buildResolveAttempts(opts gradle.ResolveOptions, flags ResolveFlags) []resolveAttempt {
	attempts := []resolveAttempt{{Options: opts, Label: "default"}}
	if strings.TrimSpace(flags.Config) != "" {
		return attempts
	}
	switch opts.Scope {
	case "compile":
		debugPattern := "*DebugCompileClasspath"
		attempt := opts
		attempt.Configs = []string{debugPattern}
		attempts = append(attempts, resolveAttempt{
			Options:        attempt,
			Label:          "config:" + debugPattern,
			ConfigPatterns: []string{debugPattern},
		})
	case "runtime":
		debugPattern := "*DebugRuntimeClasspath"
		attempt := opts
		attempt.Configs = []string{debugPattern}
		attempts = append(attempts, resolveAttempt{
			Options:        attempt,
			Label:          "config:" + debugPattern,
			ConfigPatterns: []string{debugPattern},
		})
	}
	return attempts
}

func joinHints(parts ...string) string {
	var out []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, " ")
}

func projectHint(flags ResolveFlags, meta ResolveMeta) string {
	hints := DetectProjectHints(flags.Project)
	var parts []string

	if hints.HasIncludeBuilds {
		parts = append(parts, fmt.Sprintf("Composite build detected; try: --project %s", hints.IncludeBuildHint))
	}
	if hints.Android && !metaHasConfig(meta, "*DebugCompileClasspath") && strings.TrimSpace(flags.Config) == "" {
		parts = append(parts, "Android detected; try: --config \"*DebugCompileClasspath\" (or --config debugCompileClasspath)")
	}
	if hints.KMP && strings.TrimSpace(flags.Targets) == "" {
		parts = append(parts, "KMP detected; try: --targets jvm (or another target)")
	}
	if len(parts) == 0 {
		return ""
	}
	parts = append(parts, "If resolution is slow: narrow with --subproject, --config, --targets, or --scope.")
	return strings.Join(parts, " ")
}

func metaHasConfig(meta ResolveMeta, pattern string) bool {
	for _, tried := range meta.TriedConfigPatterns {
		if tried == pattern {
			return true
		}
	}
	return false
}

package gradle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/respawn-app/ksrc/internal/executil"
	"github.com/respawn-app/ksrc/internal/resolve"
)

type ResolveOptions struct {
	ProjectDir      string
	ProjectPath     string
	Module          string
	Group           string
	Artifact        string
	Version         string
	Scope           string
	Configs         []string
	Targets         []string
	Subprojects     []string
	Dep             string
	Offline         bool
	Refresh         bool
	IncludeBuildSrc bool
}

type ResolveResult struct {
	Sources []resolve.SourceJar
	Deps    []resolve.Coord
}

func Resolve(ctx context.Context, runner executil.Runner, opts ResolveOptions) (ResolveResult, error) {
	result, err := resolveOnce(ctx, runner, opts)
	if err != nil {
		return ResolveResult{}, err
	}

	if !opts.IncludeBuildSrc {
		return result, nil
	}

	buildSrcDir := filepath.Join(opts.ProjectDir, "buildSrc")
	if shouldResolveBuildSrc(buildSrcDir, opts.ProjectDir, opts.ProjectPath) {
		buildSrcOpts := opts
		buildSrcOpts.ProjectPath = buildSrcDir
		buildSrcOpts.Subprojects = nil
		buildSrcRes, err := resolveOnce(ctx, runner, buildSrcOpts)
		if err != nil {
			return ResolveResult{}, err
		}
		result = mergeResults(result, buildSrcRes)
	}
	return result, nil
}

func resolveOnce(ctx context.Context, runner executil.Runner, opts ResolveOptions) (ResolveResult, error) {
	scriptPath, cleanup, err := writeInitScript()
	if err != nil {
		return ResolveResult{}, err
	}
	defer cleanup()

	gradleCmd, err := findGradle(runner, opts.ProjectDir)
	if err != nil {
		return ResolveResult{}, err
	}

	args := []string{"-I", scriptPath, "-Dorg.gradle.console=plain", "--info", "--no-configuration-cache"}
	if opts.ProjectPath != "" {
		args = append(args, "-p", opts.ProjectPath)
	}
	if opts.Offline {
		args = append(args, "--offline")
	}
	if opts.Refresh {
		args = append(args, "--refresh-dependencies")
	}
	args = append(args, buildProps(opts)...) // -P...
	args = append(args, "ksrcSources")

	stdout, stderr, err := runner.Run(ctx, opts.ProjectDir, gradleCmd, args...)
	if err != nil {
		return ResolveResult{}, fmt.Errorf("gradle failed: %w\n%s", err, strings.TrimSpace(stderr))
	}

	result := ResolveResult{}
	seen := make(map[string]struct{})
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "KSRC|") {
			coord, path, ok := parseLine(line, "KSRC|")
			if !ok {
				continue
			}
			key := coord.String() + "|" + path
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result.Sources = append(result.Sources, resolve.SourceJar{Coord: coord, Path: path})
			continue
		}
		if strings.HasPrefix(line, "KSRCDEP|") {
			coord, _, ok := parseLine(line, "KSRCDEP|")
			if !ok {
				continue
			}
			result.Deps = append(result.Deps, coord)
		}
	}
	return result, nil
}

func shouldResolveBuildSrc(buildSrcDir string, projectDir string, projectPath string) bool {
	if projectDir != "" && samePath(buildSrcDir, projectDir) {
		return false
	}
	if projectPath != "" && samePath(buildSrcDir, projectPath) {
		return false
	}
	info, err := os.Stat(buildSrcDir)
	if err != nil || !info.IsDir() {
		return false
	}
	if hasGradleBuildFile(buildSrcDir) {
		return true
	}
	return false
}

func hasGradleBuildFile(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "build.gradle.kts")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "settings.gradle")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "settings.gradle.kts")); err == nil {
		return true
	}
	return false
}

func samePath(a string, b string) bool {
	aAbs, errA := filepath.Abs(a)
	bAbs, errB := filepath.Abs(b)
	if errA == nil && errB == nil {
		return aAbs == bAbs
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func mergeResults(base ResolveResult, extra ResolveResult) ResolveResult {
	if len(extra.Sources) == 0 && len(extra.Deps) == 0 {
		return base
	}
	seenSources := make(map[string]struct{}, len(base.Sources))
	for _, s := range base.Sources {
		seenSources[s.Coord.String()+"|"+s.Path] = struct{}{}
	}
	for _, s := range extra.Sources {
		key := s.Coord.String() + "|" + s.Path
		if _, ok := seenSources[key]; ok {
			continue
		}
		seenSources[key] = struct{}{}
		base.Sources = append(base.Sources, s)
	}

	seenDeps := make(map[string]struct{}, len(base.Deps))
	for _, d := range base.Deps {
		seenDeps[d.String()] = struct{}{}
	}
	for _, d := range extra.Deps {
		key := d.String()
		if _, ok := seenDeps[key]; ok {
			continue
		}
		seenDeps[key] = struct{}{}
		base.Deps = append(base.Deps, d)
	}
	return base
}

func buildProps(opts ResolveOptions) []string {
	props := []string{}
	add := func(k, v string) {
		if strings.TrimSpace(v) == "" {
			return
		}
		props = append(props, "-P"+k+"="+v)
	}
	add("ksrcModule", opts.Module)
	add("ksrcGroup", opts.Group)
	add("ksrcArtifact", opts.Artifact)
	add("ksrcVersion", opts.Version)
	add("ksrcScope", opts.Scope)
	if len(opts.Configs) > 0 {
		add("ksrcConfig", strings.Join(opts.Configs, ","))
	}
	if len(opts.Targets) > 0 {
		add("ksrcTargets", strings.Join(opts.Targets, ","))
	}
	if len(opts.Subprojects) > 0 {
		add("ksrcSubprojects", strings.Join(opts.Subprojects, ","))
	}
	add("ksrcDep", opts.Dep)
	return props
}

func parseLine(line, prefix string) (resolve.Coord, string, bool) {
	trim := strings.TrimPrefix(line, prefix)
	parts := strings.SplitN(trim, "|", 2)
	coord, err := resolve.ParseCoord(parts[0])
	if err != nil {
		return resolve.Coord{}, "", false
	}
	if len(parts) == 1 {
		return coord, "", true
	}
	return coord, strings.TrimSpace(parts[1]), true
}

func findGradle(runner executil.Runner, projectDir string) (string, error) {
	wrapper := filepath.Join(projectDir, "gradlew")
	if info, err := os.Stat(wrapper); err == nil && !info.IsDir() {
		return "./gradlew", nil
	}
	path, err := runner.LookPath("gradle")
	if err == nil && path != "" {
		return "gradle", nil
	}
	return "", fmt.Errorf("gradle not found (no ./gradlew and gradle not on PATH)")
}

func writeInitScript() (string, func(), error) {
	file, err := os.CreateTemp("", "ksrc-init-*.gradle")
	if err != nil {
		return "", nil, err
	}
	if _, err := file.WriteString(InitScript()); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return "", nil, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", nil, err
	}
	cleanup := func() {
		_ = os.Remove(file.Name())
	}
	return file.Name(), cleanup, nil
}

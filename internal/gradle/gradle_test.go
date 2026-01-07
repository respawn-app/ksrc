package gradle

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/respawn-app/ksrc/internal/resolve"
)

type fakeRunner struct{}

func (fakeRunner) Run(_ context.Context, _ string, _ string, _ ...string) (string, string, error) {
	return "", "", nil
}

func (fakeRunner) LookPath(_ string) (string, error) {
	return "", errors.New("not found")
}

func TestFindGradlePrefersLocalWrapper(t *testing.T) {
	dir := t.TempDir()
	wrapper := filepath.Join(dir, "gradlew")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}
	cmd, err := findGradle(fakeRunner{}, dir, "")
	if err != nil {
		t.Fatalf("findGradle: %v", err)
	}
	if cmd != "./gradlew" {
		t.Fatalf("expected local wrapper, got %q", cmd)
	}
}

func TestFindGradleFallsBackToRootWrapper(t *testing.T) {
	root := t.TempDir()
	included := t.TempDir()
	wrapper := filepath.Join(root, "gradlew")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}
	cmd, err := findGradle(fakeRunner{}, included, root)
	if err != nil {
		t.Fatalf("findGradle: %v", err)
	}
	if cmd != wrapper {
		t.Fatalf("expected root wrapper %q, got %q", wrapper, cmd)
	}
}

func TestMergeResultsIncludesWarnings(t *testing.T) {
	base := ResolveResult{
		Sources: []resolve.SourceJar{},
		Deps:    []resolve.Coord{},
		Warnings: []string{
			"base warning",
		},
	}
	extra := ResolveResult{
		Warnings: []string{
			"extra warning",
		},
	}
	merged := mergeResults(base, extra)
	if len(merged.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(merged.Warnings))
	}
}

func TestResolveBuildSkipsBuildSrcForExplicitDep(t *testing.T) {
	dir := t.TempDir()
	buildSrcDir := filepath.Join(dir, "buildSrc")
	if err := os.MkdirAll(buildSrcDir, 0o755); err != nil {
		t.Fatalf("mkdir buildSrc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(buildSrcDir, "build.gradle.kts"), []byte(""), 0o644); err != nil {
		t.Fatalf("write buildSrc build file: %v", err)
	}

	runner := &countingRunner{}
	opts := ResolveOptions{
		ProjectDir:      dir,
		IncludeBuildSrc: true,
		Dep:             "com.example:demo:1.0.0",
	}
	if _, err := resolveBuild(context.Background(), runner, opts); err != nil {
		t.Fatalf("resolveBuild: %v", err)
	}
	if runner.runCount != 1 {
		t.Fatalf("expected 1 Gradle run, got %d", runner.runCount)
	}
}

type countingRunner struct {
	runCount int
}

func (r *countingRunner) Run(_ context.Context, _ string, _ string, _ ...string) (string, string, error) {
	r.runCount++
	return "", "", nil
}

func (r *countingRunner) LookPath(_ string) (string, error) {
	return "gradle", nil
}

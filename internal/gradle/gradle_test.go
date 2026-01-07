package gradle

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
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

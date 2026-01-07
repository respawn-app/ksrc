package cli

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ProjectHints struct {
	Android          bool
	KMP              bool
	KotlinJVM        bool
	HasIncludeBuilds bool
	IncludeBuildHint string
}

func DetectProjectHints(projectDir string) ProjectHints {
	hints := ProjectHints{}
	_, content := readSettings(projectDir)
	if content != "" {
		includeBuilds := parseIncludeBuilds(content)
		if len(includeBuilds) > 0 {
			hints.HasIncludeBuilds = true
			hints.IncludeBuildHint = filepath.Clean(filepath.Join(projectDir, includeBuilds[0]))
		}
	}
	android, kmp, kotlinJVM := detectPlugins(projectDir, 200, 64*1024)
	hints.Android = android
	hints.KMP = kmp
	hints.KotlinJVM = kotlinJVM
	return hints
}

func readSettings(projectDir string) (string, string) {
	candidates := []string{
		filepath.Join(projectDir, "settings.gradle.kts"),
		filepath.Join(projectDir, "settings.gradle"),
	}
	for _, path := range candidates {
		data, err := readFilePrefix(path, 256*1024)
		if err == nil && data != "" {
			return path, data
		}
	}
	return "", ""
}

func parseIncludeBuilds(content string) []string {
	re := regexp.MustCompile(`includeBuild\s*\(\s*[^"']*["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(content, -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		path := strings.TrimSpace(match[1])
		if path == "" {
			continue
		}
		out = append(out, path)
	}
	return out
}

func detectPlugins(projectDir string, maxFiles int, maxBytes int64) (bool, bool, bool) {
	var android bool
	var kmp bool
	var kotlinJVM bool
	visited := 0
	stopWalk := errors.New("stop-walk")

	_ = filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".gradle" || d.Name() == "build" {
				return filepath.SkipDir
			}
			if depth(path, projectDir) > 4 {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "build.gradle" && d.Name() != "build.gradle.kts" {
			return nil
		}
		visited++
		if visited > maxFiles {
			return stopWalk
		}
		data, err := readFilePrefix(path, maxBytes)
		if err != nil || data == "" {
			return nil
		}
		if !android && (strings.Contains(data, "com.android.application") ||
			strings.Contains(data, "com.android.library") ||
			strings.Contains(data, "com.android.test") ||
			strings.Contains(data, "com.android.dynamic-feature") ||
			strings.Contains(data, "com.android.feature")) {
			android = true
		}
		if !kmp && (strings.Contains(data, "org.jetbrains.kotlin.multiplatform") ||
			strings.Contains(data, "kotlin(\"multiplatform\")") ||
			strings.Contains(data, "kotlin(\"multiplatform\"")) {
			kmp = true
		}
		if !kotlinJVM && (strings.Contains(data, "org.jetbrains.kotlin.jvm") ||
			strings.Contains(data, "kotlin(\"jvm\")") ||
			strings.Contains(data, "kotlin(\"jvm\"")) {
			kotlinJVM = true
		}
		if android && kmp && kotlinJVM {
			return stopWalk
		}
		return nil
	})

	return android, kmp, kotlinJVM
}

func readFilePrefix(path string, maxBytes int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	size := info.Size()
	if size > maxBytes {
		size = maxBytes
	}
	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) && n == 0 {
		return "", err
	}
	return string(buf[:n]), nil
}

func depth(path string, root string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 0
	}
	rel = filepath.Clean(rel)
	if rel == "." {
		return 0
	}
	return len(strings.Split(rel, string(filepath.Separator)))
}

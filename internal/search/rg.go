package search

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/respawn-app/ksrc/internal/executil"
	"github.com/respawn-app/ksrc/internal/resolve"
)

type Match struct {
	FileID string
	File   string
	Line   int
	Column int
	Text   string
}

type Options struct {
	Pattern string
	Jars    []resolve.SourceJar
	RGArgs  []string
	WorkDir string
}

func Run(ctx context.Context, runner executil.Runner, opts Options) ([]Match, error) {
	if opts.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	if len(opts.Jars) == 0 {
		return nil, fmt.Errorf("no source jars to search")
	}
	if _, err := runner.LookPath("rg"); err != nil {
		return nil, fmt.Errorf("rg not found on PATH")
	}

	root, err := os.MkdirTemp("", "ksrc-search-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)

	extractRoots := make(map[string]resolve.Coord)
	searchDirs := make([]string, 0, len(opts.Jars))
	for i, j := range opts.Jars {
		dir := filepath.Join(root, fmt.Sprintf("jar-%d", i))
		if err := extractJar(j.Path, dir); err != nil {
			return nil, err
		}
		extractRoots[dir] = j.Coord
		searchDirs = append(searchDirs, dir)
	}

	args := []string{"--no-heading", "--line-number", "--column", "--color=never", "--with-filename", "-g", "*.kt"}
	args = append(args, opts.RGArgs...)
	args = append(args, opts.Pattern)
	args = append(args, searchDirs...)

	stdout, stderr, err := runner.Run(ctx, opts.WorkDir, "rg", args...)
	if err != nil {
		if strings.TrimSpace(stdout) == "" {
			return nil, fmt.Errorf("rg failed: %w\n%s", err, strings.TrimSpace(stderr))
		}
	}

	matches := []Match{}
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m, ok := parseRgLine(line)
		if !ok {
			continue
		}
		coord, inner, ok := mapToCoord(extractRoots, m.File)
		if !ok {
			continue
		}
		m.FileID = coord.String() + "!/" + inner
		matches = append(matches, m)
	}
	return matches, nil
}

func parseRgLine(line string) (Match, bool) {
	// file:line:col:match
	last := strings.LastIndex(line, ":")
	if last <= 0 {
		return Match{}, false
	}
	second := strings.LastIndex(line[:last], ":")
	if second <= 0 {
		return Match{}, false
	}
	third := strings.LastIndex(line[:second], ":")
	if third <= 0 {
		return Match{}, false
	}
	file := line[:third]
	lineStr := line[third+1 : second]
	colStr := line[second+1 : last]
	text := line[last+1:]
	ln, err := strconv.Atoi(lineStr)
	if err != nil {
		return Match{}, false
	}
	col, err := strconv.Atoi(colStr)
	if err != nil {
		return Match{}, false
	}
	return Match{File: file, Line: ln, Column: col, Text: text}, true
}

func mapToCoord(roots map[string]resolve.Coord, filePath string) (resolve.Coord, string, bool) {
	for root, coord := range roots {
		rel, err := filepath.Rel(root, filePath)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		rel = filepath.ToSlash(rel)
		return coord, rel, true
	}
	return resolve.Coord{}, "", false
}

func extractJar(src, dest string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		path := filepath.Join(dest, filepath.FromSlash(f.Name))
		clean := filepath.Clean(path)
		if !strings.HasPrefix(clean, dest) {
			return fmt.Errorf("invalid path in archive: %s", f.Name)
		}
		if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(clean)
		if err != nil {
			_ = rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			_ = out.Close()
			_ = rc.Close()
			return err
		}
		_ = out.Close()
		_ = rc.Close()
	}
	return nil
}

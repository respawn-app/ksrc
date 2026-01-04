package search

import "testing"

func TestParseRgLine(t *testing.T) {
	line := "/tmp/lib.jar:com/foo/Bar.kt:12:3:match text"
	m, ok := parseRgLine(line)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if m.File != "/tmp/lib.jar:com/foo/Bar.kt" || m.Line != 12 || m.Column != 3 || m.Text != "match text" {
		t.Fatalf("unexpected match: %+v", m)
	}
}

func TestParseRgContextLine(t *testing.T) {
	line := "/tmp/foo-bar/baz.kt-7-context line"
	m, ok := parseRgLine(line)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if m.File != "/tmp/foo-bar/baz.kt" || m.Line != 7 || m.Column != 0 || m.Text != "context line" {
		t.Fatalf("unexpected match: %+v", m)
	}
}

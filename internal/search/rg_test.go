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

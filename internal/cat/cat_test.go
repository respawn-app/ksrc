package cat

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestParseLineRange(t *testing.T) {
	lr, err := ParseLineRange("1,3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lr.Start != 1 || lr.End != 3 {
		t.Fatalf("unexpected range: %+v", lr)
	}
	if _, err := ParseLineRange("1:3"); err == nil {
		t.Fatal("expected error for invalid range")
	}
}

func TestReadFileFromZipRange(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "test.jar")
	inner := "a/b.txt"

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create(inner)
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	_, _ = w.Write([]byte("line1\nline2\nline3\n"))
	_ = zw.Close()
	_ = f.Close()

	lr := &LineRange{Start: 2, End: 3}
	data, err := ReadFileFromZip(zipPath, inner, lr)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "line2\nline3\n" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}

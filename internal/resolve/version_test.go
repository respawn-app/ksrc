package resolve

import "testing"

func TestCompareVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.1", "1.0.0", 1},
		{"2.0", "10.0", -1},
		{"1.2.3", "1.2.3", 0},
		{"1.2.0", "1.1.9", 1},
		{"1.2", "1.2.0", 0},
	}
	for _, c := range cases {
		if got := CompareVersion(c.a, c.b); sign(got) != c.want {
			t.Fatalf("CompareVersion(%q,%q)=%d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

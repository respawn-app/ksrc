package resolve

import (
	"strconv"
	"strings"
	"unicode"
)

type versionToken struct {
	num bool
	val string
}

func tokenizeVersion(s string) []versionToken {
	var tokens []versionToken
	var buf strings.Builder
	var mode rune // 'n' or 's'
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		tokens = append(tokens, versionToken{num: mode == 'n', val: buf.String()})
		buf.Reset()
	}
	for _, r := range s {
		if r == '.' || r == '-' || r == '_' || r == '+' {
			flush()
			mode = 0
			continue
		}
		if unicode.IsDigit(r) {
			if mode != 'n' {
				flush()
				mode = 'n'
			}
			buf.WriteRune(r)
			continue
		}
		if unicode.IsLetter(r) {
			if mode != 's' {
				flush()
				mode = 's'
			}
			buf.WriteRune(unicode.ToLower(r))
			continue
		}
		// treat other chars as separators
		flush()
		mode = 0
	}
	flush()
	return tokens
}

// CompareVersion compares two Maven-like versions.
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func CompareVersion(a, b string) int {
	at := trimTokens(tokenizeVersion(a))
	bt := trimTokens(tokenizeVersion(b))
	max := len(at)
	if len(bt) > max {
		max = len(bt)
	}
	for i := 0; i < max; i++ {
		var av, bv versionToken
		if i < len(at) {
			av = at[i]
		}
		if i < len(bt) {
			bv = bt[i]
		}
		if av.num && bv.num {
			ai, _ := strconv.Atoi(av.val)
			bi, _ := strconv.Atoi(bv.val)
			if ai < bi {
				return -1
			}
			if ai > bi {
				return 1
			}
			continue
		}
		if av.num && !bv.num {
			if av.val == "" && bv.val == "" {
				continue
			}
			return 1
		}
		if !av.num && bv.num {
			if av.val == "" && bv.val == "" {
				continue
			}
			return -1
		}
		// both string or empty
		if av.val < bv.val {
			return -1
		}
		if av.val > bv.val {
			return 1
		}
	}
	return 0
}

func trimTokens(tokens []versionToken) []versionToken {
	i := len(tokens)
	for i > 0 {
		t := tokens[i-1]
		if t.num && isZeroToken(t.val) {
			i--
			continue
		}
		if !t.num && t.val == "" {
			i--
			continue
		}
		break
	}
	return tokens[:i]
}

func isZeroToken(v string) bool {
	for _, r := range v {
		if r != '0' {
			return false
		}
	}
	return v != ""
}

package resolve

import (
	"path"
	"strings"
)

// MatchAny checks if value matches any of the comma-separated glob patterns.
func MatchAny(patterns string, value string) bool {
	if patterns == "" {
		return true
	}
	for _, p := range strings.Split(patterns, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		ok, err := path.Match(p, value)
		if err == nil && ok {
			return true
		}
		if strings.EqualFold(p, value) {
			return true
		}
	}
	return false
}

// MatchModule matches a module selector against a coordinate.
// The selector may be:
// - group:artifact[:version]
// - a loose token that can match group, artifact, or group:artifact
func MatchModule(selector string, c Coord) bool {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return true
	}
	if strings.Contains(selector, ":") {
		parts := strings.Split(selector, ":")
		if len(parts) >= 2 {
			if !MatchAny(parts[0], c.Group) {
				return false
			}
			if !MatchAny(parts[1], c.Artifact) {
				return false
			}
			if len(parts) >= 3 && parts[2] != "" {
				return MatchAny(parts[2], c.Version)
			}
			return true
		}
	}

	candidates := []string{
		c.Group,
		c.Artifact,
		c.Group + ":" + c.Artifact,
		normalizeDots(c.Artifact),
		normalizeDots(c.Group + ":" + c.Artifact),
	}
	for _, cand := range candidates {
		if MatchAny(selector, cand) {
			return true
		}
		if strings.Contains(cand, selector) {
			return true
		}
	}
	return false
}

func normalizeDots(s string) string {
	return strings.ReplaceAll(s, "-", ".")
}

// FilterSources applies module/group/artifact/version filters.
func FilterSources(sources []SourceJar, module, group, artifact, version string) []SourceJar {
	out := make([]SourceJar, 0, len(sources))
	for _, s := range sources {
		if module != "" && !MatchModule(module, s.Coord) {
			continue
		}
		if !MatchAny(group, s.Coord.Group) {
			continue
		}
		if !MatchAny(artifact, s.Coord.Artifact) {
			continue
		}
		if !MatchAny(version, s.Coord.Version) {
			continue
		}
		out = append(out, s)
	}
	return out
}

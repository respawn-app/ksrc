package resolve

import "strings"

// SelectorToCoord attempts to extract an exact group/artifact[/version] from flags.
func SelectorToCoord(module, group, artifact, version string) (Coord, bool) {
	if group != "" && artifact != "" {
		return Coord{Group: group, Artifact: artifact, Version: version}, true
	}
	module = strings.TrimSpace(module)
	if strings.Contains(module, ":") {
		parts := strings.Split(module, ":")
		if len(parts) >= 2 {
			c := Coord{Group: parts[0], Artifact: parts[1]}
			if len(parts) >= 3 {
				c.Version = parts[2]
			}
			if version != "" {
				c.Version = version
			}
			if c.Group != "" && c.Artifact != "" {
				return c, true
			}
		}
	}
	return Coord{}, false
}

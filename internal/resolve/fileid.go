package resolve

import (
	"fmt"
	"strings"
)

// ParseFileID parses group:artifact:version!/path/inside.jar
func ParseFileID(value string) (Coord, string, error) {
	parts := strings.SplitN(value, "!/", 2)
	if len(parts) != 2 {
		return Coord{}, "", fmt.Errorf("invalid file-id: %q", value)
	}
	coord, err := ParseCoord(parts[0])
	if err != nil {
		return Coord{}, "", err
	}
	path := strings.TrimPrefix(parts[1], "/")
	if path == "" {
		return Coord{}, "", fmt.Errorf("invalid file-id: %q", value)
	}
	return coord, path, nil
}

package filebrowser

import (
	"path"
	"strings"
)

const (
	PathSeparator = "/"
)

// GetClosestRelative returns the closest path between two given paths
func GetClosestRelative(p1 string, p2 string) (index int) {
	p1Split := strings.Split(p1, PathSeparator)
	p2Split := strings.Split(p2, PathSeparator)

	for index = 0; index < len(p1Split) && index < len(p2Split); index++ {
		if p1Split[index] != p2Split[index] {
			break
		}
	}

	return
}

// NormalizePath cleans the path ensuring it is absolute
func NormalizePath(p string) string {
	if !path.IsAbs(p) {
		p = path.Join(PathSeparator, p)
	}

	return path.Clean(p)
}

func PathComponents(p string) []string {
	paths := strings.Split(p, PathSeparator)

	// if the path p starts with "/" this will supose an empty string at the
	// beginning of the paths array, something similar happens if the path p
	// ends with "/", since an extra empty string will be added at the end of
	// the array. Substracting 2 to the length ensures only 1 extra allocation
	// will be required in the worst case.
	components := make([]string, 0, len(paths)-2)

	for _, p := range paths {
		if len(p) > 0 {
			components = append(components, p)
		}
	}

	return components
}

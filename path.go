package filebrowser

import (
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

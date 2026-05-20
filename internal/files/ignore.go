package files

import "strings"

var ignoredDirectoryPrefixes = []string{".git", "vendor", "node_modules", "var", "build", "dist", "coverage"}

func IsIgnoredPath(path string) bool {
	normalized := strings.TrimPrefix(path, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	for _, prefix := range ignoredDirectoryPrefixes {
		if normalized == prefix || strings.HasPrefix(normalized, prefix+"/") || strings.Contains(normalized, "/"+prefix+"/") {
			return true
		}
	}
	for _, prefix := range []string{"storage/framework", "bootstrap/cache"} {
		if normalized == prefix || strings.HasPrefix(normalized, prefix+"/") || strings.Contains(normalized, "/"+prefix+"/") {
			return true
		}
	}
	return false
}

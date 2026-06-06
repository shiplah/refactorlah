package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FindGoRootForPaths(projectRoot string, paths []string) (string, bool, error) {
	return FindMarkerRootForPaths(projectRoot, paths, []string{"go.mod"})
}

func ReadGoModulePath(moduleRoot string) (string, error) {
	contents, err := os.ReadFile(filepath.Join(moduleRoot, "go.mod"))
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}

	return "", fmt.Errorf("go.mod at %s does not declare a module path", moduleRoot)
}

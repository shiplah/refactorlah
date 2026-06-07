package golang

import (
	"path/filepath"

	adapterproto "refactorlah/internal/adapters/contract"
)

func goSanityChecks(projectRoot string, goRoot string, available func(string) bool) []adapterproto.Check {
	if !available("go") {
		return nil
	}

	return []adapterproto.Check{{
		Directory: relativeDirectory(projectRoot, goRoot),
		Command:   []string{"go", "build", "./..."},
	}}
}

func relativeDirectory(projectRoot string, root string) string {
	relative, err := filepath.Rel(projectRoot, root)
	if err != nil || relative == "." {
		return ""
	}
	return filepath.ToSlash(relative)
}

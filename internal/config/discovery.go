package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shiplah/refactorlah/internal/files"
)

func (l *Loader) findConfigFiles(searchRoot string) ([]string, error) {
	configFiles := []string{}
	err := filepath.WalkDir(searchRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(searchRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)

		if entry.IsDir() {
			if relative != "." && files.IsIgnoredPath(relative) {
				return filepath.SkipDir
			}
			if depth(relative) > maxSearchDepth {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.Name() != fileName || depth(filepath.ToSlash(filepath.Dir(relative))) > maxSearchDepth {
			return nil
		}

		configFiles = append(configFiles, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(configFiles, func(i, j int) bool {
		left := filepath.ToSlash(configFiles[i])
		right := filepath.ToSlash(configFiles[j])
		leftDepth := strings.Count(left, "/")
		rightDepth := strings.Count(right, "/")
		if leftDepth == rightDepth {
			return left < right
		}
		return leftDepth < rightDepth
	})

	return configFiles, nil
}

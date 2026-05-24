package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const fileName = ".refactorlah.json"

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load(projectRoot string) (Config, error) {
	files, err := l.findConfigFiles(projectRoot)
	if err != nil {
		return Config{}, err
	}

	merged := Config{}
	for _, file := range files {
		config, err := readConfigFile(file)
		if err != nil {
			return Config{}, err
		}

		configDir, err := filepath.Rel(projectRoot, filepath.Dir(file))
		if err != nil {
			return Config{}, err
		}
		configDir = filepath.ToSlash(configDir)
		if configDir == "." {
			configDir = ""
		}

		merged.Include = append(merged.Include, qualifyPatterns(configDir, config.Include)...)
		merged.Exclude = append(merged.Exclude, qualifyPatterns(configDir, config.Exclude)...)
	}

	return merged, nil
}

func (l *Loader) findConfigFiles(projectRoot string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(projectRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)

		if entry.IsDir() && relative != "." && ignoredDir(relative) {
			return filepath.SkipDir
		}

		if entry.IsDir() || entry.Name() != fileName {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		left := filepath.ToSlash(files[i])
		right := filepath.ToSlash(files[j])
		leftDepth := strings.Count(left, "/")
		rightDepth := strings.Count(right, "/")
		if leftDepth == rightDepth {
			return left < right
		}
		return leftDepth < rightDepth
	})

	return files, nil
}

func readConfigFile(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var payload struct {
		Include []string `json:"include"`
		Exclude []string `json:"exclude"`
	}
	if err := json.Unmarshal(content, &payload); err != nil {
		return Config{}, err
	}

	return Config{
		Include: nonEmptyStrings(payload.Include),
		Exclude: nonEmptyStrings(payload.Exclude),
	}, nil
}

func qualifyPatterns(configDir string, patterns []string) []string {
	qualified := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		if pattern == "" {
			continue
		}
		if strings.HasPrefix(pattern, "/") {
			qualified = append(qualified, strings.TrimPrefix(pattern, "/"))
			continue
		}
		if configDir == "" {
			qualified = append(qualified, pattern)
			continue
		}
		qualified = append(qualified, configDir+"/"+pattern)
	}
	return qualified
}

func nonEmptyStrings(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(filepath.ToSlash(item))
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func ignoredDir(path string) bool {
	for _, prefix := range []string{
		".git",
		"vendor",
		"node_modules",
		"var",
		"storage/framework",
		"bootstrap/cache",
		"build",
		"dist",
		"coverage",
	} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}

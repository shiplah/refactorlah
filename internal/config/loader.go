package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"refactorlah/internal/files"
)

const (
	fileName       = ".refactorlah.json"
	maxSearchDepth = 3
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load(projectRoot string, searchRoot string) (Config, error) {
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return Config{}, err
	}
	absSearchRoot, err := filepath.Abs(searchRoot)
	if err != nil {
		return Config{}, err
	}

	if ok, err := isWithin(absProjectRoot, absSearchRoot); err != nil {
		return Config{}, err
	} else if !ok {
		return Config{}, fmt.Errorf("config search root %q is outside project root %q", searchRoot, projectRoot)
	}

	files, err := l.findConfigFiles(absSearchRoot)
	if err != nil {
		return Config{}, err
	}

	index := newPatternIndex(absProjectRoot)
	for _, file := range files {
		config, err := readConfigFile(file)
		if err != nil {
			return Config{}, err
		}
		if err := index.addIncludes(filepath.Dir(file), config.Include); err != nil {
			return Config{}, err
		}
		if err := index.addExcludes(filepath.Dir(file), config.Exclude); err != nil {
			return Config{}, err
		}
	}

	return index.config(), nil
}

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

func depth(path string) int {
	path = strings.Trim(filepath.ToSlash(path), "/")
	if path == "" || path == "." {
		return 0
	}
	return strings.Count(path, "/") + 1
}

func isWithin(root string, candidate string) (bool, error) {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false, err
	}
	relative = filepath.ToSlash(relative)
	return relative == "." || (!strings.HasPrefix(relative, "../") && relative != ".."), nil
}

type patternIndex struct {
	projectRoot string
	include     []string
	exclude     []string
	seenInclude map[string]struct{}
	seenExclude map[string]struct{}
}

func newPatternIndex(projectRoot string) *patternIndex {
	return &patternIndex{
		projectRoot: projectRoot,
		seenInclude: map[string]struct{}{},
		seenExclude: map[string]struct{}{},
	}
}

func (i *patternIndex) addIncludes(configDir string, patterns []string) error {
	return i.addPatterns(configDir, patterns, &i.include, i.seenInclude)
}

func (i *patternIndex) addExcludes(configDir string, patterns []string) error {
	return i.addPatterns(configDir, patterns, &i.exclude, i.seenExclude)
}

func (i *patternIndex) addPatterns(configDir string, patterns []string, target *[]string, seen map[string]struct{}) error {
	for _, pattern := range patterns {
		absolute, relative, err := i.normalizePattern(configDir, pattern)
		if err != nil {
			return err
		}
		if absolute == "" {
			continue
		}
		if _, ok := seen[absolute]; ok {
			continue
		}
		seen[absolute] = struct{}{}
		*target = append(*target, relative)
	}

	return nil
}

func (i *patternIndex) normalizePattern(configDir string, pattern string) (string, string, error) {
	pattern = strings.TrimSpace(filepath.ToSlash(pattern))
	if pattern == "" {
		return "", "", nil
	}

	var absolute string
	if strings.HasPrefix(pattern, "/") {
		absolute = filepath.Join(i.projectRoot, filepath.FromSlash(strings.TrimPrefix(pattern, "/")))
	} else {
		absolute = filepath.Join(configDir, filepath.FromSlash(pattern))
	}
	absolute = filepath.Clean(absolute)

	inside, err := isWithin(i.projectRoot, absolute)
	if err != nil {
		return "", "", err
	}
	if !inside {
		return "", "", fmt.Errorf("config pattern %q resolves outside project root", pattern)
	}

	relative, err := filepath.Rel(i.projectRoot, absolute)
	if err != nil {
		return "", "", err
	}

	absolute = filepath.ToSlash(absolute)
	relative = filepath.ToSlash(relative)
	if relative == "." {
		relative = ""
	}

	return absolute, relative, nil
}

func (i *patternIndex) config() Config {
	return Config{
		Include: append([]string(nil), i.include...),
		Exclude: append([]string(nil), i.exclude...),
	}
}

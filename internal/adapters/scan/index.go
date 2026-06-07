package scan

import (
	"fmt"
	"path/filepath"
	"sort"

	"refactorlah/internal/config"
	"refactorlah/internal/files"
)

type collectorFunc func(root string, relativePath string) ([]string, error)

type Index struct {
	projectRoot string
	scanConfig  config.Config
	collector   collectorFunc
	cache       map[string][]string
}

func NewIndex(projectRoot string, scanConfig config.Config) *Index {
	return newIndex(projectRoot, scanConfig, files.CollectFiles)
}

func newIndex(projectRoot string, scanConfig config.Config, collector collectorFunc) *Index {
	absoluteProjectRoot, err := filepath.Abs(projectRoot)
	if err == nil {
		projectRoot = absoluteProjectRoot
	}

	return &Index{
		projectRoot: filepath.Clean(projectRoot),
		scanConfig:  scanConfig,
		collector:   collector,
		cache:       map[string][]string{},
	}
}

func (i *Index) Files(root string, extensions ...string) ([]string, error) {
	allFiles, err := i.filesInRoot(root)
	if err != nil {
		return nil, err
	}
	if len(extensions) == 0 {
		return copyStrings(allFiles), nil
	}

	wanted := map[string]bool{}
	for _, extension := range extensions {
		wanted[extension] = true
	}

	var selected []string
	for _, file := range allFiles {
		if wanted[filepath.Ext(file)] {
			selected = append(selected, file)
		}
	}
	return selected, nil
}

func (i *Index) filesInRoot(root string) ([]string, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absoluteRoot = filepath.Clean(absoluteRoot)

	if cached, ok := i.cache[absoluteRoot]; ok {
		return cached, nil
	}

	rootRelativeToProject, err := filepath.Rel(i.projectRoot, absoluteRoot)
	if err != nil {
		return nil, err
	}
	if rootRelativeToProject == ".." || filepath.IsAbs(rootRelativeToProject) || startsWithParentTraversal(rootRelativeToProject) {
		return nil, fmt.Errorf("scan root %q is outside project root %q", root, i.projectRoot)
	}

	collected, err := i.collector(absoluteRoot, ".")
	if err != nil {
		return nil, err
	}

	projectRelativeFiles := make([]string, 0, len(collected))
	for _, rootRelativeFile := range collected {
		projectRelative := rootRelativeFile
		if rootRelativeToProject != "." {
			projectRelative = filepath.Join(rootRelativeToProject, filepath.FromSlash(rootRelativeFile))
		}
		projectRelative = filepath.ToSlash(projectRelative)
		if !i.scanConfig.Allows(projectRelative) {
			continue
		}
		projectRelativeFiles = append(projectRelativeFiles, projectRelative)
	}
	sort.Strings(projectRelativeFiles)

	i.cache[absoluteRoot] = projectRelativeFiles
	return projectRelativeFiles, nil
}

func startsWithParentTraversal(path string) bool {
	return len(path) > 3 && path[:3] == ".."+string(filepath.Separator)
}

func copyStrings(values []string) []string {
	copied := make([]string, len(values))
	copy(copied, values)
	return copied
}

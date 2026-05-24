package config

import (
	"fmt"
	"path/filepath"
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
	absProjectRoot, err := resolveExistingPath(projectRoot)
	if err != nil {
		return Config{}, err
	}
	absSearchRoot, err := resolveExistingPath(searchRoot)
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

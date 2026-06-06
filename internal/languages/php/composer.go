package php

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type composerConfig struct {
	Autoload    composerAutoload `json:"autoload"`
	AutoloadDev composerAutoload `json:"autoload-dev"`
}

type composerAutoload struct {
	Psr4 map[string]composerPsr4Paths `json:"psr-4"`
}

type composerPsr4Paths []string

func (p *composerPsr4Paths) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*p = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err != nil {
		return err
	}
	*p = multiple
	return nil
}

func ReadComposerPsr4Map(projectRoot string, composerRoot string) (Psr4Map, error) {
	content, err := os.ReadFile(filepath.Join(composerRoot, "composer.json"))
	if err != nil {
		return Psr4Map{}, fmt.Errorf("read composer.json: %w", err)
	}

	var config composerConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return Psr4Map{}, fmt.Errorf("parse composer.json: %w", err)
	}

	prefix, err := filepath.Rel(projectRoot, composerRoot)
	if err != nil {
		return Psr4Map{}, err
	}
	prefix = filepath.ToSlash(prefix)
	if prefix == "." {
		prefix = ""
	}

	mappings := map[string][]string{}
	appendMappings(mappings, prefix, config.Autoload.Psr4)
	appendMappings(mappings, prefix, config.AutoloadDev.Psr4)

	return NewPsr4Map(mappings), nil
}

func appendMappings(target map[string][]string, prefix string, mappings map[string]composerPsr4Paths) {
	for namespace, paths := range mappings {
		for _, composerPath := range paths {
			normalized := normalizeComposerPath(prefix, composerPath)
			target[namespace] = append(target[namespace], normalized)
		}
	}
}

func normalizeComposerPath(prefix string, composerPath string) string {
	normalized := strings.Trim(strings.ReplaceAll(composerPath, "\\", "/"), "/")
	if normalized == "" {
		normalized = "."
	}
	if prefix == "" {
		return normalized
	}
	if normalized == "." {
		return prefix
	}
	return path.Join(prefix, normalized)
}

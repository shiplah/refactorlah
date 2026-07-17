//go:build cgo

package php

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
)

func ReadComposerAutoloadFiles(projectRoot string, composerRoot string) (map[string]bool, error) {
	content, err := os.ReadFile(filepath.Join(composerRoot, "composer.json"))
	if err != nil {
		return nil, fmt.Errorf("read composer.json: %w", err)
	}

	var config composerConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("parse composer.json: %w", err)
	}

	prefix, err := filepath.Rel(projectRoot, composerRoot)
	if err != nil {
		return nil, err
	}
	prefix = filepath.ToSlash(prefix)
	if prefix == "." {
		prefix = ""
	}

	files := map[string]bool{}
	for _, composerPath := range append(config.Autoload.Files, config.AutoloadDev.Files...) {
		files[normalizeComposerPath(prefix, composerPath)] = true
	}
	return files, nil
}

func CollectComposerAutoloadFileReplacements(projectRoot string, composerRoot string, plan planning.MovePlan) ([]adapterproto.Replacement, error) {
	composerFile := filepath.Join(composerRoot, "composer.json")
	content, err := os.ReadFile(composerFile)
	if err != nil {
		return nil, fmt.Errorf("read composer.json: %w", err)
	}

	var config composerConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("parse composer.json: %w", err)
	}

	prefix, err := filepath.Rel(projectRoot, composerRoot)
	if err != nil {
		return nil, err
	}
	prefix = filepath.ToSlash(prefix)
	if prefix == "." {
		prefix = ""
	}

	composerRelativeFile, err := filepath.Rel(projectRoot, composerFile)
	if err != nil {
		return nil, err
	}
	composerRelativeFile = filepath.ToSlash(composerRelativeFile)

	var result []adapterproto.Replacement
	for _, composerPath := range append(config.Autoload.Files, config.AutoloadDev.Files...) {
		normalisedPath := normalizeComposerPath(prefix, composerPath)
		for _, move := range plan.Moves {
			if normalisedPath != move.OldPath {
				continue
			}

			newComposerPath, err := relativeComposerPath(projectRoot, composerRoot, move.NewPath, composerPath)
			if err != nil {
				return nil, err
			}
			start, end, ok := jsonStringContentRange(content, composerPath)
			if !ok {
				continue
			}
			result = append(result, adapterproto.Replacement{
				File:        composerRelativeFile,
				Start:       start,
				End:         end,
				Replacement: escapedJSONStringContent(newComposerPath),
				Reason:      "php-composer-autoload-files",
				Adapter:     "php",
			})
		}
	}

	return result, nil
}

func relativeComposerPath(projectRoot string, composerRoot string, projectRelativePath string, previousComposerPath string) (string, error) {
	absoluteNewPath := filepath.Join(projectRoot, filepath.FromSlash(projectRelativePath))
	relative, err := filepath.Rel(composerRoot, absoluteNewPath)
	if err != nil {
		return "", err
	}
	relative = filepath.ToSlash(relative)
	if strings.HasPrefix(previousComposerPath, "./") && !strings.HasPrefix(relative, "../") {
		return "./" + relative, nil
	}
	return relative, nil
}

func jsonStringContentRange(content []byte, value string) (int, int, bool) {
	for index := 0; index < len(content); index++ {
		if content[index] != '"' {
			continue
		}
		end := index + 1
		escaped := false
		for end < len(content) {
			if content[end] == '"' && !escaped {
				break
			}
			escaped = content[end] == '\\' && !escaped
			if content[end] != '\\' {
				escaped = false
			}
			end++
		}
		if end >= len(content) {
			return 0, 0, false
		}

		unquoted, err := strconv.Unquote(string(content[index : end+1]))
		if err == nil && unquoted == value {
			return index + 1, end, true
		}
		index = end
	}

	return 0, 0, false
}

func escapedJSONStringContent(value string) string {
	quoted := strconv.Quote(value)
	return quoted[1 : len(quoted)-1]
}

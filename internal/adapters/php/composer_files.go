//go:build cgo

package php

import (
	"path/filepath"
	"strconv"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
)

func ReadComposerAutoloadFiles(projectRoot string, composerRoot string) (map[string]bool, error) {
	composer, err := readComposerData(projectRoot, composerRoot)
	if err != nil {
		return nil, err
	}

	files := map[string]bool{}
	for _, composerPath := range composer.autoloadFiles() {
		files[normalizeComposerPath(composer.prefix, composerPath)] = true
	}
	return files, nil
}

func CollectComposerAutoloadFileReplacements(projectRoot string, composerRoot string, plan planning.MovePlan) ([]adapterproto.Replacement, error) {
	composer, err := readComposerData(projectRoot, composerRoot)
	if err != nil {
		return nil, err
	}

	composerRelativeFile, err := filepath.Rel(projectRoot, composer.file)
	if err != nil {
		return nil, err
	}
	composerRelativeFile = filepath.ToSlash(composerRelativeFile)

	var result []adapterproto.Replacement
	for _, composerPath := range composer.autoloadFiles() {
		normalisedPath := normalizeComposerPath(composer.prefix, composerPath)
		for _, move := range plan.Moves {
			if normalisedPath != move.OldPath {
				continue
			}

			newComposerPath, err := relativeComposerPath(projectRoot, composerRoot, move.NewPath, composerPath)
			if err != nil {
				return nil, err
			}
			start, end, ok := jsonStringContentRange(composer.source, composerPath)
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

func (c composerData) autoloadFiles() []string {
	files := make([]string, 0, len(c.config.Autoload.Files)+len(c.config.AutoloadDev.Files))
	files = append(files, c.config.Autoload.Files...)
	files = append(files, c.config.AutoloadDev.Files...)
	return files
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

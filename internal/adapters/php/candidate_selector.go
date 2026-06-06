//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
)

type CandidateFileSelector struct{}

func (s CandidateFileSelector) Select(projectRoot string, files []string, mappings []adapterproto.SymbolMapping) []string {
	if len(mappings) == 0 {
		return nil
	}

	movedFiles := map[string]bool{}
	for _, mapping := range mappings {
		movedFiles[mapping.OldPath] = true
	}

	needles := candidateNeedles(mappings)
	var selected []string
	for _, file := range files {
		if movedFiles[file] {
			selected = append(selected, file)
			continue
		}

		content, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
		if err != nil || len(content) == 0 {
			continue
		}

		if containsAnyNeedle(string(content), needles) {
			selected = append(selected, file)
		}
	}

	sort.Strings(selected)
	return selected
}

func candidateNeedles(mappings []adapterproto.SymbolMapping) []string {
	index := map[string]bool{}
	for _, mapping := range mappings {
		index[mapping.OldSymbol] = true
		index[mapping.OldNamespace] = true
		index[shortSymbolName(mapping.OldSymbol)] = true
		for needle := range literalHints(mapping) {
			index[needle] = true
		}
	}

	needles := make([]string, 0, len(index))
	for needle := range index {
		if needle != "" {
			needles = append(needles, needle)
		}
	}
	sort.Strings(needles)
	return needles
}

func containsAnyNeedle(content string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(content, needle) {
			return true
		}
	}
	return false
}

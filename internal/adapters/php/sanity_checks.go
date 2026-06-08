//go:build cgo

package php

import (
	"path/filepath"
	"sort"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/planning"
)

func phpSanityChecks(projectRoot string, composerRoot string, plan planning.MovePlan, replacements []adapterproto.Replacement, available func(string) bool) []adapterproto.Check {
	checks := []adapterproto.Check{}

	if available("php") {
		for _, file := range editedPHPFiles(plan, replacements) {
			checks = append(checks, adapterproto.Check{
				Command: []string{"php", "-l", file},
			})
		}
	}

	if available("composer") {
		checks = append(checks, adapterproto.Check{
			Directory: relativeDirectory(projectRoot, composerRoot),
			Command:   []string{"composer", "dump-autoload"},
		})
	}

	return checks
}

func editedPHPFiles(plan planning.MovePlan, replacements []adapterproto.Replacement) []string {
	files := map[string]struct{}{}
	movedPaths := plan.TargetPaths()
	for _, move := range plan.Moves {
		if filepath.Ext(move.NewPath) == ".php" {
			files[move.NewPath] = struct{}{}
		}
	}
	for _, replacement := range replacements {
		file := replacement.File
		if moved, ok := movedPaths[file]; ok {
			file = moved
		}
		if filepath.Ext(file) == ".php" {
			files[file] = struct{}{}
		}
	}

	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}
	sort.Strings(result)
	return result
}

func relativeDirectory(projectRoot string, root string) string {
	relative, err := filepath.Rel(projectRoot, root)
	if err != nil || relative == "." {
		return ""
	}
	return filepath.ToSlash(relative)
}

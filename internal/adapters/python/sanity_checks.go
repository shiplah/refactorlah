//go:build cgo

package python

import (
	"path/filepath"
	"sort"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/planning"
)

func pythonSanityChecks(plan planning.MovePlan, replacements []adapterproto.Replacement, commandPath func(string) (string, bool)) []adapterproto.Check {
	python, ok := pythonCommand(commandPath)
	if !ok {
		return nil
	}

	files := editedPythonFiles(plan, replacements)
	checks := make([]adapterproto.Check, 0, len(files))
	for _, file := range files {
		checks = append(checks, adapterproto.Check{
			Command: []string{python, "-m", "py_compile", file},
		})
	}
	return checks
}

func pythonCommand(commandPath func(string) (string, bool)) (string, bool) {
	if path, ok := commandPath("python3"); ok {
		return path, true
	}
	return commandPath("python")
}

func editedPythonFiles(plan planning.MovePlan, replacements []adapterproto.Replacement) []string {
	files := map[string]struct{}{}
	movedPaths := plan.TargetPaths()
	for _, move := range plan.Moves {
		if filepath.Ext(move.NewPath) == ".py" {
			files[move.NewPath] = struct{}{}
		}
	}
	for _, replacement := range replacements {
		file := replacement.File
		if moved, ok := movedPaths[file]; ok {
			file = moved
		}
		if filepath.Ext(file) == ".py" {
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

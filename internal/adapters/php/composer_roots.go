//go:build cgo

package php

import (
	"path/filepath"

	"refactorlah/internal/adapters/shared"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
)

func composerRootsForPlan(projectRoot string, plan planning.MovePlan) ([]string, bool, error) {
	paths := phpRelevantMovePaths(plan)
	if len(paths) == 0 {
		return nil, false, nil
	}
	return project.FindComposerRootsForPaths(projectRoot, paths)
}

func phpRelevantMovePaths(plan planning.MovePlan) []string {
	if plan.IsDir {
		return shared.MovePaths(plan)
	}

	var paths []string
	for _, move := range plan.Moves {
		if isPHPRelevantExtension(filepath.Ext(move.OldPath)) || isPHPRelevantExtension(filepath.Ext(move.NewPath)) {
			paths = append(paths, move.OldPath, move.NewPath)
		}
	}
	return paths
}

func isPHPRelevantExtension(extension string) bool {
	switch extension {
	case ".php", ".twig", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".css":
		return true
	default:
		return false
	}
}

func planForComposerRoot(projectRoot string, composerRoot string, plan planning.MovePlan) planning.MovePlan {
	rootPlan := plan
	rootPlan.Moves = nil
	for _, move := range plan.Moves {
		if pathBelongsToRoot(projectRoot, composerRoot, move.OldPath) || pathBelongsToRoot(projectRoot, composerRoot, move.NewPath) {
			rootPlan.Moves = append(rootPlan.Moves, move)
		}
	}
	rootPlan.IsDir = plan.IsDir && len(rootPlan.Moves) > 0
	return rootPlan
}

func pathBelongsToRoot(projectRoot string, root string, relativePath string) bool {
	absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	relativeToRoot, err := filepath.Rel(root, absolutePath)
	if err != nil {
		return false
	}
	return relativeToRoot == "." || relativeToRoot != ".." && !filepath.IsAbs(relativeToRoot) && !startsWithParent(relativeToRoot)
}

func startsWithParent(path string) bool {
	return len(path) > 3 && path[:3] == ".."+string(filepath.Separator)
}

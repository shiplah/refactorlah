package core

import (
	"strings"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/planning"
)

func ProjectDirectoryPathMappings(plan planning.MovePlan) []adapterproto.PathMapping {
	if !plan.IsDir {
		return nil
	}

	oldPath := strings.TrimRight(plan.OldPath, "/")
	newPath := strings.TrimRight(plan.NewPath, "/")
	if oldPath == "" || newPath == "" || oldPath == newPath {
		return nil
	}

	return []adapterproto.PathMapping{{
		Kind:         "project-path-directory",
		OldPath:      oldPath,
		NewPath:      newPath,
		OldReference: oldPath + "/",
		NewReference: newPath + "/",
	}}
}

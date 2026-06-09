package core

import (
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
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

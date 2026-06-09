package twig

import (
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
)

type TemplateMapper struct{}

func (m TemplateMapper) DeriveMappings(moves []planning.FileMove, configuration PathConfiguration) []adapterproto.PathMapping {
	var mappings []adapterproto.PathMapping
	seenDirectories := map[string]bool{}

	for _, move := range moves {
		if !strings.HasSuffix(move.OldPath, ".twig") || !strings.HasSuffix(move.NewPath, ".twig") {
			continue
		}

		oldReference, ok := referenceForPath(move.OldPath, configuration)
		if !ok {
			continue
		}
		newReference, ok := referenceForPath(move.NewPath, configuration)
		if !ok {
			continue
		}

		mappings = append(mappings, adapterproto.PathMapping{
			Kind:         "twig-template",
			OldPath:      move.OldPath,
			NewPath:      move.NewPath,
			OldReference: oldReference,
			NewReference: newReference,
		})

		oldDirectoryReference, oldOK := directoryReference(oldReference)
		newDirectoryReference, newOK := directoryReference(newReference)
		directoryKey := oldDirectoryReference + "\x00" + newDirectoryReference
		if oldOK && newOK && oldDirectoryReference != newDirectoryReference && !seenDirectories[directoryKey] {
			mappings = append(mappings, adapterproto.PathMapping{
				Kind:         "twig-template-directory",
				OldPath:      move.OldPath,
				NewPath:      move.NewPath,
				OldReference: oldDirectoryReference,
				NewReference: newDirectoryReference,
			})
			seenDirectories[directoryKey] = true
		}
	}

	return mappings
}

func referenceForPath(path string, configuration PathConfiguration) (string, bool) {
	var bestRoot *PathRoot
	for index := range configuration.Roots {
		root := configuration.Roots[index]
		if path != root.Path && !strings.HasPrefix(path, root.Path+"/") {
			continue
		}
		if bestRoot == nil || len(root.Path) > len(bestRoot.Path) {
			bestRoot = &root
		}
	}
	if bestRoot == nil {
		return "", false
	}

	relative := strings.TrimPrefix(path, bestRoot.Path)
	relative = strings.TrimPrefix(relative, "/")
	if bestRoot.Namespace == "" {
		return relative, true
	}

	return "@" + bestRoot.Namespace + "/" + relative, true
}

func directoryReference(reference string) (string, bool) {
	index := strings.LastIndex(reference, "/")
	if index <= 0 {
		return "", false
	}

	return reference[:index], true
}

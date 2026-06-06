package php

import (
	"path"
	"sort"
	"strings"
)

type Psr4Map struct {
	mappings map[string][]string
}

func NewPsr4Map(mappings map[string][]string) Psr4Map {
	normalised := make(map[string][]string, len(mappings))
	for namespace, paths := range mappings {
		copied := make([]string, 0, len(paths))
		for _, pathValue := range paths {
			copied = append(copied, strings.Trim(strings.ReplaceAll(pathValue, "\\", "/"), "/"))
		}
		normalised[namespace] = copied
	}

	return Psr4Map{mappings: normalised}
}

type ResolvedSymbol struct {
	Symbol    string
	Namespace string
	ShortName string
}

type Psr4NamespaceResolver struct{}

func (r Psr4NamespaceResolver) DeriveSymbol(psr4 Psr4Map, relativePath string) (ResolvedSymbol, bool) {
	if path.Ext(relativePath) != ".php" {
		return ResolvedSymbol{}, false
	}

	normalisedPath := strings.TrimPrefix(strings.ReplaceAll(relativePath, "\\", "/"), "/")
	type candidate struct {
		namespace string
		basePath  string
	}
	var candidates []candidate

	for namespace, paths := range psr4.mappings {
		for _, basePath := range paths {
			if basePath == "" || basePath == "." {
				candidates = append(candidates, candidate{namespace: namespace, basePath: "."})
				continue
			}
			if normalisedPath == basePath || strings.HasPrefix(normalisedPath, basePath+"/") {
				candidates = append(candidates, candidate{namespace: namespace, basePath: basePath})
			}
		}
	}

	if len(candidates) == 0 {
		return ResolvedSymbol{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		return len(candidates[i].basePath) > len(candidates[j].basePath)
	})

	best := candidates[0]
	relative := normalisedPath
	if best.basePath != "." {
		relative = strings.TrimPrefix(normalisedPath, best.basePath+"/")
	}
	relative = strings.TrimSuffix(relative, ".php")
	if relative == "" {
		return ResolvedSymbol{}, false
	}

	parts := strings.Split(relative, "/")
	shortName := parts[len(parts)-1]
	namespace := strings.TrimSuffix(best.namespace, "\\")
	if len(parts) > 1 {
		namespace += "\\" + strings.Join(parts[:len(parts)-1], "\\")
	}

	return ResolvedSymbol{
		Symbol:    namespace + "\\" + shortName,
		Namespace: namespace,
		ShortName: shortName,
	}, true
}

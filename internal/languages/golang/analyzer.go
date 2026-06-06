package golang

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/files"
	"refactorlah/internal/languages"
	"refactorlah/internal/languages/golang/rules"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
)

type Analyzer struct {
	importRule rules.ImportPathRule
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{importRule: rules.ImportPathRule{}}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	if !plan.ContainsExtension(".go") {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	goRoot, found, err := project.FindGoRootForPaths(projectRoot, languages.MovePaths(plan))
	if err != nil || !found {
		return adapterproto.AggregatedResponse{}, found, err
	}

	modulePath, err := project.ReadGoModulePath(goRoot)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}

	mappings, err := importPathMappings(projectRoot, goRoot, modulePath, plan)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	if len(mappings) == 0 {
		return adapterproto.AggregatedResponse{}, true, nil
	}

	replacements, warnings, err := a.collectImportReplacements(projectRoot, goRoot, mappings)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}

	pathMappings := make([]adapterproto.PathMapping, 0, len(mappings))
	for _, mapping := range mappings {
		pathMappings = append(pathMappings, adapterproto.PathMapping{
			Kind:         "go-import-path",
			OldPath:      mapping.OldPath,
			NewPath:      mapping.NewPath,
			OldReference: mapping.OldImport,
			NewReference: mapping.NewImport,
		})
	}

	return adapterproto.AggregatedResponse{
		PathMappings: pathMappings,
		Replacements: replacements,
		Warnings:     warnings,
	}, true, nil
}

type importPathMapping struct {
	OldPath   string
	NewPath   string
	OldImport string
	NewImport string
}

func importPathMappings(projectRoot string, goRoot string, modulePath string, plan planning.MovePlan) ([]importPathMapping, error) {
	seen := map[string]bool{}
	var mappings []importPathMapping

	for _, move := range plan.Moves {
		if filepath.Ext(move.OldPath) != ".go" && filepath.Ext(move.NewPath) != ".go" {
			continue
		}

		oldDirectory := path.Dir(move.OldPath)
		newDirectory := path.Dir(move.NewPath)
		if oldDirectory == newDirectory {
			continue
		}

		oldImport, err := importPathForDirectory(projectRoot, goRoot, modulePath, oldDirectory)
		if err != nil {
			return nil, err
		}
		newImport, err := importPathForDirectory(projectRoot, goRoot, modulePath, newDirectory)
		if err != nil {
			return nil, err
		}

		key := oldImport + "\x00" + newImport
		if seen[key] {
			continue
		}
		seen[key] = true

		mappings = append(mappings, importPathMapping{
			OldPath:   oldDirectory,
			NewPath:   newDirectory,
			OldImport: oldImport,
			NewImport: newImport,
		})
	}

	return mappings, nil
}

func importPathForDirectory(projectRoot string, goRoot string, modulePath string, directory string) (string, error) {
	absoluteDirectory := filepath.Join(projectRoot, filepath.FromSlash(directory))
	relativeDirectory, err := filepath.Rel(goRoot, absoluteDirectory)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(relativeDirectory, "..") {
		return "", fmt.Errorf("go package directory %s is outside module root %s", directory, goRoot)
	}

	relativeSlash := filepath.ToSlash(relativeDirectory)
	if relativeSlash == "." {
		return modulePath, nil
	}

	return modulePath + "/" + relativeSlash, nil
}

func (a *Analyzer) collectImportReplacements(projectRoot string, goRoot string, mappings []importPathMapping) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
	goFiles, err := files.CollectFiles(goRoot, ".")
	if err != nil {
		return nil, nil, err
	}

	ruleMappings := make([]rules.ImportPathMapping, 0, len(mappings))
	for _, mapping := range mappings {
		ruleMappings = append(ruleMappings, rules.ImportPathMapping{
			OldImport: mapping.OldImport,
			NewImport: mapping.NewImport,
		})
	}

	var replacements []adapterproto.Replacement
	var warnings []adapterproto.Warning
	for _, goFile := range goFiles {
		if filepath.Ext(goFile) != ".go" {
			continue
		}

		absolutePath := filepath.Join(goRoot, filepath.FromSlash(goFile))
		source, err := os.ReadFile(absolutePath)
		if err != nil {
			return nil, nil, err
		}

		projectRelativePath, err := filepath.Rel(projectRoot, absolutePath)
		if err != nil {
			return nil, nil, err
		}
		projectRelativePath = filepath.ToSlash(projectRelativePath)

		fileReplacements, err := a.importRule.Collect(source, rules.ImportPathInput{
			File:     projectRelativePath,
			Mappings: ruleMappings,
		})
		if err != nil {
			warnings = append(warnings, adapterproto.Warning{
				File:    projectRelativePath,
				Message: fmt.Sprintf("Go imports not analysed because file could not be parsed: %v", err),
			})
			continue
		}
		replacements = append(replacements, fileReplacements...)
	}

	return replacements, warnings, nil
}

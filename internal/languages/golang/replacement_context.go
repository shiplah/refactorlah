package golang

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/languages/golang/rules"
)

func packageDeclarationMappingsByFile(mappings []packageMoveMapping) map[string]filePackageMapping {
	result := map[string]filePackageMapping{}
	for _, mapping := range mappings {
		for _, filePackage := range mapping.FilePackages {
			if filePackage.OldPackage == filePackage.NewPackage {
				continue
			}
			result[filePackage.OldPath] = filePackage
		}
	}
	return result
}

func symbolDeclarationMappingsByFile(mappings []symbolMoveMapping) map[string][]rules.SymbolDeclarationMapping {
	result := map[string][]rules.SymbolDeclarationMapping{}
	for _, mapping := range mappings {
		result[mapping.OldPath] = append(result[mapping.OldPath], rules.SymbolDeclarationMapping{
			File:      mapping.OldPath,
			OldSymbol: mapping.OldSymbol,
			NewSymbol: mapping.NewSymbol,
			Kind:      mapping.Kind,
		})
	}
	return result
}

func importedSymbolReferenceMappings(mappings []symbolMoveMapping) []rules.ImportedSymbolReferenceMapping {
	result := make([]rules.ImportedSymbolReferenceMapping, 0, len(mappings))
	for _, mapping := range mappings {
		result = append(result, rules.ImportedSymbolReferenceMapping{
			OldImport:  mapping.OldImport,
			NewImport:  mapping.NewImport,
			OldPackage: mapping.OldPackage,
			NewPackage: mapping.NewPackage,
			OldSymbol:  mapping.OldSymbol,
			NewSymbol:  mapping.NewSymbol,
		})
	}
	return result
}

func (a *Analyzer) collectLocalSymbolReferences(projectRoot string, mappings []symbolMoveMapping) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
	groups := localSymbolMappingGroups(mappings)
	var replacements []adapterproto.Replacement
	var warnings []adapterproto.Warning

	for _, key := range sortedLocalSymbolGroupKeys(groups) {
		group := groups[key]
		files, err := goSourceFilesInDirectory(projectRoot, group.directory)
		if err != nil {
			return nil, nil, err
		}

		fileReplacements, err := a.localSymbolRule.Collect(rules.LocalSymbolReferenceInput{
			PackageName: group.packageName,
			Files:       files,
			Mappings:    group.mappings,
		})
		if err != nil {
			warnings = append(warnings, adapterproto.Warning{
				File:    group.directory,
				Message: fmt.Sprintf("Go local symbol references not analysed because package could not be parsed: %v", err),
			})
			continue
		}
		replacements = append(replacements, fileReplacements...)
	}

	return replacements, warnings, nil
}

type localSymbolMappingGroup struct {
	directory   string
	packageName string
	mappings    []rules.LocalSymbolReferenceMapping
}

func localSymbolMappingGroups(mappings []symbolMoveMapping) map[string]localSymbolMappingGroup {
	groups := map[string]localSymbolMappingGroup{}
	for _, mapping := range mappings {
		directory := path.Dir(mapping.OldPath)
		key := directory + "\x00" + mapping.OldPackage
		group := groups[key]
		group.directory = directory
		group.packageName = mapping.OldPackage
		group.mappings = append(group.mappings, rules.LocalSymbolReferenceMapping{
			File:      mapping.OldPath,
			OldSymbol: mapping.OldSymbol,
			NewSymbol: mapping.NewSymbol,
		})
		groups[key] = group
	}
	return groups
}

func sortedLocalSymbolGroupKeys(groups map[string]localSymbolMappingGroup) []string {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func goSourceFilesInDirectory(projectRoot string, directory string) ([]rules.GoSourceFile, error) {
	goFiles, err := goFilesInDirectory(projectRoot, directory)
	if err != nil {
		return nil, err
	}

	sourceFiles := make([]rules.GoSourceFile, 0, len(goFiles))
	for _, goFile := range goFiles {
		source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(goFile)))
		if err != nil {
			return nil, err
		}
		sourceFiles = append(sourceFiles, rules.GoSourceFile{
			File:   goFile,
			Source: source,
		})
	}
	return sourceFiles, nil
}

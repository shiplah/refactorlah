package golang

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/files"
	"refactorlah/internal/languages"
	"refactorlah/internal/languages/golang/rules"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
)

type Analyzer struct {
	importRule             rules.ImportPathRule
	packageDeclarationRule rules.PackageDeclarationRule
	packageQualifierRule   rules.PackageQualifierRule
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		importRule:             rules.ImportPathRule{},
		packageDeclarationRule: rules.PackageDeclarationRule{},
		packageQualifierRule:   rules.PackageQualifierRule{},
	}
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

	mappings, mappingWarnings, err := packageMoveMappings(projectRoot, goRoot, modulePath, plan)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	if len(mappings) == 0 {
		return adapterproto.AggregatedResponse{Warnings: mappingWarnings}, true, nil
	}

	replacements, warnings, err := a.collectReplacements(projectRoot, goRoot, mappings)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	warnings = append(mappingWarnings, warnings...)

	pathMappings := make([]adapterproto.PathMapping, 0, len(mappings))
	symbolMappings := make([]adapterproto.SymbolMapping, 0, len(mappings))
	for _, mapping := range mappings {
		pathMappings = append(pathMappings, adapterproto.PathMapping{
			Kind:         "go-import-path",
			OldPath:      mapping.OldPath,
			NewPath:      mapping.NewPath,
			OldReference: mapping.OldImport,
			NewReference: mapping.NewImport,
		})
		symbolMappings = append(symbolMappings, adapterproto.SymbolMapping{
			Kind:         "package",
			OldPath:      mapping.OldPath,
			NewPath:      mapping.NewPath,
			OldSymbol:    mapping.OldImport,
			NewSymbol:    mapping.NewImport,
			OldNamespace: path.Dir(mapping.OldImport),
			NewNamespace: path.Dir(mapping.NewImport),
			ShortName:    mapping.OldPackage,
		})
	}

	return adapterproto.AggregatedResponse{
		SymbolMappings: symbolMappings,
		PathMappings:   pathMappings,
		Replacements:   replacements,
		Warnings:       warnings,
	}, true, nil
}

func (a *Analyzer) collectReplacements(projectRoot string, goRoot string, mappings []packageMoveMapping) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
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
	packageQualifierMappings := make([]rules.PackageQualifierMapping, 0, len(mappings))
	for _, mapping := range mappings {
		packageQualifierMappings = append(packageQualifierMappings, rules.PackageQualifierMapping{
			OldImport:  mapping.OldImport,
			NewImport:  mapping.NewImport,
			OldPackage: mapping.OldPackage,
			NewPackage: mapping.NewPackage,
		})
	}
	packageDeclarationMappings := packageDeclarationMappingsByFile(mappings)

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

		fileReplacements, err = a.packageQualifierRule.Collect(source, rules.PackageQualifierInput{
			File:     projectRelativePath,
			Mappings: packageQualifierMappings,
		})
		if err != nil {
			warnings = append(warnings, adapterproto.Warning{
				File:    projectRelativePath,
				Message: fmt.Sprintf("Go package qualifiers not analysed because file could not be parsed: %v", err),
			})
			continue
		}
		replacements = append(replacements, fileReplacements...)

		if packageMapping, ok := packageDeclarationMappings[projectRelativePath]; ok {
			fileReplacements, err = a.packageDeclarationRule.Collect(source, rules.PackageDeclarationInput{
				File:       projectRelativePath,
				OldPackage: packageMapping.OldPackage,
				NewPackage: packageMapping.NewPackage,
			})
			if err != nil {
				warnings = append(warnings, adapterproto.Warning{
					File:    projectRelativePath,
					Message: fmt.Sprintf("Go package declaration not analysed because file could not be parsed: %v", err),
				})
				continue
			}
			replacements = append(replacements, fileReplacements...)
		}
	}

	return replacements, warnings, nil
}

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

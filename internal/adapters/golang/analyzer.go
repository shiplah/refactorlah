package golang

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/golang/rules"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/adapters/shared"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
)

type Analyzer struct {
	importRule             rules.ImportPathRule
	importedSymbolRule     rules.ImportedSymbolReferenceRule
	localSymbolRule        rules.LocalSymbolReferenceRule
	packageDeclarationRule rules.PackageDeclarationRule
	packageQualifierRule   rules.PackageQualifierRule
	symbolDeclarationRule  rules.SymbolDeclarationRule
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		importRule:             rules.ImportPathRule{},
		importedSymbolRule:     rules.ImportedSymbolReferenceRule{},
		localSymbolRule:        rules.LocalSymbolReferenceRule{},
		packageDeclarationRule: rules.PackageDeclarationRule{},
		packageQualifierRule:   rules.PackageQualifierRule{},
		symbolDeclarationRule:  rules.SymbolDeclarationRule{},
	}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (adapterproto.AggregatedResponse, bool, error) {
	_ = scanConfig

	if !plan.ContainsExtension(".go") {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	goRoot, found, err := project.FindGoRootForPaths(projectRoot, shared.MovePaths(plan))
	if err != nil || !found {
		return adapterproto.AggregatedResponse{}, found, err
	}

	modulePath, err := project.ReadGoModulePath(goRoot)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}

	packageMappings, mappingWarnings, err := packageMoveMappings(projectRoot, goRoot, modulePath, plan)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	symbolMappings, symbolWarnings, err := symbolMoveMappings(projectRoot, goRoot, modulePath, plan, packageMappings)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	mappingWarnings = append(mappingWarnings, symbolWarnings...)
	if len(packageMappings) == 0 && len(symbolMappings) == 0 {
		return adapterproto.AggregatedResponse{Warnings: mappingWarnings}, true, nil
	}

	replacements, warnings, err := a.collectReplacements(projectRoot, goRoot, packageMappings, symbolMappings, scanIndex)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	warnings = append(mappingWarnings, warnings...)

	pathMappings := make([]adapterproto.PathMapping, 0, len(packageMappings))
	responseSymbolMappings := make([]adapterproto.SymbolMapping, 0, len(packageMappings)+len(symbolMappings))
	for _, mapping := range packageMappings {
		pathMappings = append(pathMappings, adapterproto.PathMapping{
			Kind:         "go-import-path",
			OldPath:      mapping.OldPath,
			NewPath:      mapping.NewPath,
			OldReference: mapping.OldImport,
			NewReference: mapping.NewImport,
		})
		responseSymbolMappings = append(responseSymbolMappings, adapterproto.SymbolMapping{
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
	for _, mapping := range symbolMappings {
		responseSymbolMappings = append(responseSymbolMappings, adapterproto.SymbolMapping{
			Kind:         "go-" + mapping.Kind,
			OldPath:      mapping.OldPath,
			NewPath:      mapping.NewPath,
			OldSymbol:    mapping.OldImport + "." + mapping.OldSymbol,
			NewSymbol:    mapping.NewImport + "." + mapping.NewSymbol,
			OldNamespace: mapping.OldImport,
			NewNamespace: mapping.NewImport,
			ShortName:    mapping.OldSymbol,
		})
	}

	return adapterproto.AggregatedResponse{
		SymbolMappings: responseSymbolMappings,
		PathMappings:   pathMappings,
		Replacements:   replacements,
		Warnings:       warnings,
	}, true, nil
}

func (a *Analyzer) collectReplacements(projectRoot string, goRoot string, packageMappings []packageMoveMapping, symbolMappings []symbolMoveMapping, scanIndex *scan.Index) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
	goFiles, err := scanIndex.Files(goRoot, ".go")
	if err != nil {
		return nil, nil, err
	}

	ruleMappings := make([]rules.ImportPathMapping, 0, len(packageMappings))
	for _, mapping := range packageMappings {
		ruleMappings = append(ruleMappings, rules.ImportPathMapping{
			OldImport: mapping.OldImport,
			NewImport: mapping.NewImport,
		})
	}
	packageQualifierMappings := make([]rules.PackageQualifierMapping, 0, len(packageMappings))
	for _, mapping := range packageMappings {
		packageQualifierMappings = append(packageQualifierMappings, rules.PackageQualifierMapping{
			OldImport:  mapping.OldImport,
			NewImport:  mapping.NewImport,
			OldPackage: mapping.OldPackage,
			NewPackage: mapping.NewPackage,
		})
	}
	packageDeclarationMappings := packageDeclarationMappingsByFile(packageMappings)
	importedSymbolMappings := importedSymbolReferenceMappings(symbolMappings)
	symbolDeclarationMappings := symbolDeclarationMappingsByFile(symbolMappings)

	var replacements []adapterproto.Replacement
	var warnings []adapterproto.Warning
	localReplacements, localWarnings, err := a.collectLocalSymbolReferences(projectRoot, symbolMappings)
	if err != nil {
		return nil, nil, err
	}
	replacements = append(replacements, localReplacements...)
	warnings = append(warnings, localWarnings...)

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

		fileReplacements, err = a.importedSymbolRule.Collect(source, rules.ImportedSymbolReferenceInput{
			File:     projectRelativePath,
			Mappings: importedSymbolMappings,
		})
		if err != nil {
			warnings = append(warnings, adapterproto.Warning{
				File:    projectRelativePath,
				Message: fmt.Sprintf("Go imported symbol references not analysed because file could not be parsed: %v", err),
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

		if symbolMappings, ok := symbolDeclarationMappings[projectRelativePath]; ok {
			fileReplacements, err = a.symbolDeclarationRule.Collect(source, rules.SymbolDeclarationInput{
				File:     projectRelativePath,
				Mappings: symbolMappings,
			})
			if err != nil {
				warnings = append(warnings, adapterproto.Warning{
					File:    projectRelativePath,
					Message: fmt.Sprintf("Go symbol declarations not analysed because file could not be parsed: %v", err),
				})
				continue
			}
			replacements = append(replacements, fileReplacements...)
		}
	}

	return replacements, warnings, nil
}

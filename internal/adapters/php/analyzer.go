//go:build cgo

package php

import (
	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/adapters/shared"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
)

type Analyzer struct {
	symbolScanner            *SymbolScanner
	referenceCollector       ReferenceCollector
	yamlSymbolCollector      YamlSymbolCollector
	semanticWarningCollector SemanticWarningCollector
	twigCollector            TwigCollector
	projectPathCollector     ProjectPathCollector
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbolScanner:            NewSymbolScanner(),
		referenceCollector:       NewReferenceCollector(),
		yamlSymbolCollector:      NewYamlSymbolCollector(),
		semanticWarningCollector: NewSemanticWarningCollector(),
		twigCollector:            NewTwigCollector(),
		projectPathCollector:     NewProjectPathCollector(),
	}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (adapterproto.AggregatedResponse, bool, error) {
	_ = scanConfig

	containsPHP := plan.ContainsExtension(".php")
	containsTwig := plan.ContainsExtension(".twig")
	containsStaticImport := planContainsStaticImportTarget(plan)
	containsProjectDirectoryPath := plan.IsDir
	if !containsPHP && !containsTwig && !containsStaticImport && !containsProjectDirectoryPath {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	composerRoot, found, err := project.FindComposerRootForPaths(projectRoot, shared.MovePaths(plan))
	if err != nil || !found {
		return adapterproto.AggregatedResponse{}, found, err
	}

	var symbolMappings []adapterproto.SymbolMapping
	var pathMappings []adapterproto.PathMapping
	var replacements []adapterproto.Replacement
	var warnings []adapterproto.Warning

	if containsPHP {
		psr4, err := ReadComposerPsr4Map(projectRoot, composerRoot)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}

		phpSymbolMappings, phpWarnings := a.symbolScanner.Scan(projectRoot, psr4, plan.Moves)
		phpReplacements, replacementWarnings, err := a.referenceCollector.Collect(projectRoot, composerRoot, phpSymbolMappings, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}
		yamlReplacements, err := a.yamlSymbolCollector.Collect(projectRoot, composerRoot, phpSymbolMappings, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}
		semanticWarnings, err := a.semanticWarningCollector.Collect(projectRoot, composerRoot, phpSymbolMappings, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}

		symbolMappings = append(symbolMappings, phpSymbolMappings...)
		replacements = append(replacements, phpReplacements...)
		replacements = append(replacements, yamlReplacements...)
		warnings = append(warnings, phpWarnings...)
		warnings = append(warnings, replacementWarnings...)
		warnings = append(warnings, semanticWarnings...)
	}

	if containsTwig {
		twigPathMappings, twigReplacements, twigWarnings, err := a.twigCollector.Collect(projectRoot, composerRoot, plan, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}
		pathMappings = append(pathMappings, twigPathMappings...)
		replacements = append(replacements, twigReplacements...)
		warnings = append(warnings, twigWarnings...)
	}

	projectPathMappings, pathReplacements, err := a.projectPathCollector.Collect(projectRoot, composerRoot, plan, containsStaticImport, scanIndex)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	pathMappings = append(pathMappings, projectPathMappings...)
	replacements = append(replacements, pathReplacements...)

	return adapterproto.AggregatedResponse{
		SymbolMappings: symbolMappings,
		PathMappings:   pathMappings,
		Replacements:   replacements,
		Warnings:       warnings,
	}, true, nil
}

func planContainsStaticImportTarget(plan planning.MovePlan) bool {
	for _, extension := range []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".css"} {
		if plan.ContainsExtension(extension) {
			return true
		}
	}
	return false
}

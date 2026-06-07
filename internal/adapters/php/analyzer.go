//go:build cgo

package php

import (
	"os/exec"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

type Analyzer struct {
	symbolScanner            *SymbolScanner
	referenceCollector       ReferenceCollector
	yamlSymbolCollector      YamlSymbolCollector
	semanticWarningCollector SemanticWarningCollector
	twigCollector            TwigCollector
	projectPathCollector     ProjectPathCollector
	commandAvailable         func(string) bool
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbolScanner:            NewSymbolScanner(),
		referenceCollector:       NewReferenceCollector(),
		yamlSymbolCollector:      NewYamlSymbolCollector(),
		semanticWarningCollector: NewSemanticWarningCollector(),
		twigCollector:            NewTwigCollector(),
		projectPathCollector:     NewProjectPathCollector(),
		commandAvailable:         commandAvailable,
	}
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
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

	composerRoots, found, err := composerRootsForPlan(projectRoot, plan)
	if err != nil || !found {
		return adapterproto.AggregatedResponse{}, found, err
	}

	var response adapterproto.AggregatedResponse
	for _, composerRoot := range composerRoots {
		rootPlan := planForComposerRoot(projectRoot, composerRoot, plan)
		if len(rootPlan.Moves) == 0 {
			continue
		}

		rootResponse, err := a.analyzeComposerRoot(projectRoot, composerRoot, rootPlan, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}
		response = appendPHPResponse(response, rootResponse)
	}

	return response, true, nil
}

func (a *Analyzer) analyzeComposerRoot(projectRoot string, composerRoot string, plan planning.MovePlan, scanIndex *scan.Index) (adapterproto.AggregatedResponse, error) {
	containsPHP := plan.ContainsExtension(".php")
	containsTwig := plan.ContainsExtension(".twig")
	containsStaticImport := planContainsStaticImportTarget(plan)

	var symbolMappings []adapterproto.SymbolMapping
	var pathMappings []adapterproto.PathMapping
	var replacements []adapterproto.Replacement
	var warnings []adapterproto.Warning

	if containsPHP {
		psr4, err := ReadComposerPsr4Map(projectRoot, composerRoot)
		if err != nil {
			return adapterproto.AggregatedResponse{}, err
		}

		phpSymbolMappings, phpWarnings := a.symbolScanner.Scan(projectRoot, psr4, plan.Moves)
		phpReplacements, replacementWarnings, err := a.referenceCollector.Collect(projectRoot, composerRoot, phpSymbolMappings, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, err
		}
		yamlReplacements, err := a.yamlSymbolCollector.Collect(projectRoot, composerRoot, phpSymbolMappings, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, err
		}
		semanticWarnings, err := a.semanticWarningCollector.Collect(projectRoot, composerRoot, phpSymbolMappings, scanIndex)
		if err != nil {
			return adapterproto.AggregatedResponse{}, err
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
			return adapterproto.AggregatedResponse{}, err
		}
		pathMappings = append(pathMappings, twigPathMappings...)
		replacements = append(replacements, twigReplacements...)
		warnings = append(warnings, twigWarnings...)
	}

	projectPathMappings, pathReplacements, err := a.projectPathCollector.Collect(projectRoot, composerRoot, plan, containsStaticImport, scanIndex)
	if err != nil {
		return adapterproto.AggregatedResponse{}, err
	}
	pathMappings = append(pathMappings, projectPathMappings...)
	replacements = append(replacements, pathReplacements...)

	return adapterproto.AggregatedResponse{
		SymbolMappings: symbolMappings,
		PathMappings:   pathMappings,
		Replacements:   replacements,
		Warnings:       warnings,
		Checks:         phpSanityChecks(projectRoot, composerRoot, plan, replacements, a.commandAvailable),
	}, nil
}

func appendPHPResponse(left adapterproto.AggregatedResponse, right adapterproto.AggregatedResponse) adapterproto.AggregatedResponse {
	left.SymbolMappings = append(left.SymbolMappings, right.SymbolMappings...)
	left.PathMappings = append(left.PathMappings, right.PathMappings...)
	left.Replacements = append(left.Replacements, right.Replacements...)
	left.Warnings = append(left.Warnings, right.Warnings...)
	left.Checks = append(left.Checks, right.Checks...)
	return left
}

func planContainsStaticImportTarget(plan planning.MovePlan) bool {
	for _, extension := range []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".css"} {
		if plan.ContainsExtension(extension) {
			return true
		}
	}
	return false
}

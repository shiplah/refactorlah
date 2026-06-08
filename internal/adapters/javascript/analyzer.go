package javascript

import (
	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/adapters/shared"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

type Analyzer struct {
	scanner staticimports.Scanner
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		scanner: staticimports.Scanner{},
	}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (adapterproto.AggregatedResponse, bool, error) {
	_ = scanConfig

	if !containsJavaScriptModuleMove(plan) {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	files, err := scanIndex.CandidateFiles(projectRoot, scan.CandidateQuery{
		Extensions:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		Needles:      moduleCandidateNeedles(plan.Moves),
		IncludePaths: shared.MovePaths(plan),
	})
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}

	replacements, err := a.collectModuleReplacements(projectRoot, files, plan.Moves)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	aliasReplacements, aliasWarnings, err := a.collectTypeScriptAliasReplacements(projectRoot, plan, scanIndex)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	replacements = append(replacements, aliasReplacements...)
	packageImportReplacements, packageWarnings, err := a.collectPackageSpecifierReplacements(projectRoot, plan, scanIndex)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	replacements = append(replacements, packageImportReplacements...)

	return adapterproto.AggregatedResponse{
		Replacements: shared.ToAdapterReplacements(replacements),
		Warnings:     append(aliasWarnings, packageWarnings...),
	}, true, nil
}

func containsJavaScriptModuleMove(plan planning.MovePlan) bool {
	for _, extension := range []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"} {
		if plan.ContainsExtension(extension) {
			return true
		}
	}
	return false
}

func (a *Analyzer) collectModuleReplacements(projectRoot string, files []string, moves []planning.FileMove) ([]replacements.Replacement, error) {
	var allReplacements []replacements.Replacement
	for _, file := range files {
		replacements, err := a.scanner.ScanSpecifiers(projectRoot, []string{file}, moduleSpecifierRewrites(file, moves))
		if err != nil {
			return nil, err
		}
		allReplacements = append(allReplacements, replacements...)
	}
	return allReplacements, nil
}

func (a *Analyzer) collectTypeScriptAliasReplacements(projectRoot string, plan planning.MovePlan, scanIndex *scan.Index) ([]replacements.Replacement, []adapterproto.Warning, error) {
	pathConfig, found, err := readTypeScriptPathConfig(projectRoot)
	if err != nil || !found {
		return nil, nil, err
	}

	warnings := typeScriptPathWarnings(projectRoot, pathConfig, plan.Moves)
	rewrites := pathAliasSpecifierRewrites(pathConfig, plan.Moves)
	configReplacements := typeScriptPathTargetReplacements(projectRoot, pathConfig, plan.Moves)
	if len(rewrites) == 0 {
		return configReplacements, warnings, nil
	}

	files, err := scanIndex.CandidateFiles(projectRoot, specifierRewriteCandidateQuery(rewrites))
	if err != nil {
		return nil, nil, err
	}
	codeReplacements, err := a.scanner.ScanSpecifiers(projectRoot, files, rewrites)
	if err != nil {
		return nil, nil, err
	}
	return append(configReplacements, codeReplacements...), warnings, nil
}

func (a *Analyzer) collectPackageSpecifierReplacements(projectRoot string, plan planning.MovePlan, scanIndex *scan.Index) ([]replacements.Replacement, []adapterproto.Warning, error) {
	packageConfig, found, err := readPackageSpecifierConfig(projectRoot)
	if err != nil || !found {
		return nil, nil, err
	}

	warnings := packageImportWarnings(packageConfig, plan.Moves)
	rewrites := packageSpecifierRewrites(packageConfig, plan.Moves)
	configReplacements := packageImportTargetReplacements(packageConfig, plan.Moves)
	if len(rewrites) == 0 {
		return configReplacements, warnings, nil
	}

	files, err := scanIndex.CandidateFiles(projectRoot, specifierRewriteCandidateQuery(rewrites))
	if err != nil {
		return nil, nil, err
	}
	codeReplacements, err := a.scanner.ScanSpecifiers(projectRoot, files, rewrites)
	if err != nil {
		return nil, nil, err
	}
	return append(configReplacements, codeReplacements...), warnings, nil
}

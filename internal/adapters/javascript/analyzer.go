package javascript

import (
	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/javascript/rules"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
	"github.com/NickSdot/refactorlah/internal/adapters/shared"
	"github.com/NickSdot/refactorlah/internal/adapters/staticimports"
	"github.com/NickSdot/refactorlah/internal/config"
	"github.com/NickSdot/refactorlah/internal/planning"
	"github.com/NickSdot/refactorlah/internal/replacements"
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
		Extensions:   rules.JavaScriptModuleExtensions(),
		Needles:      rules.ModuleCandidateNeedles(plan.Moves),
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
	bundlerReplacements, bundlerWarnings, err := a.collectBundlerAliasReplacements(projectRoot, plan, scanIndex)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	replacements = append(replacements, bundlerReplacements...)

	return adapterproto.AggregatedResponse{
		Replacements: shared.ToAdapterReplacements(replacements),
		Warnings:     append(append(aliasWarnings, packageWarnings...), bundlerWarnings...),
	}, true, nil
}

func containsJavaScriptModuleMove(plan planning.MovePlan) bool {
	for _, extension := range rules.JavaScriptModuleExtensions() {
		if plan.ContainsExtension(extension) {
			return true
		}
	}
	return false
}

func (a *Analyzer) collectModuleReplacements(projectRoot string, files []string, moves []planning.FileMove) ([]replacements.Replacement, error) {
	var allReplacements []replacements.Replacement
	rule := rules.ModuleSpecifierRule{}
	for _, file := range files {
		replacements, err := a.scanner.ScanSpecifiers(projectRoot, []string{file}, rule.Collect(file, moves))
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

	files, err := scanIndex.CandidateFiles(projectRoot, rules.SpecifierRewriteCandidateQuery(rewrites))
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

	files, err := scanIndex.CandidateFiles(projectRoot, rules.SpecifierRewriteCandidateQuery(rewrites))
	if err != nil {
		return nil, nil, err
	}
	codeReplacements, err := a.scanner.ScanSpecifiers(projectRoot, files, rewrites)
	if err != nil {
		return nil, nil, err
	}
	return append(configReplacements, codeReplacements...), warnings, nil
}

//go:build cgo

package python

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/python/rules"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/adapters/shared"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/planning"
)

type Analyzer struct {
	sourceRootResolver    SourceRootResolver
	importRule            rules.ImportStatementRule
	relativeImportRule    rules.RelativeImportRule
	importedReferenceRule rules.ImportedModuleReferenceRule
	qualifiedRule         rules.QualifiedModuleReferenceRule
	stringAnnotationRule  rules.StringAnnotationRule
	configScanner         DottedPathReferenceScanner
	commandPath           func(string) (string, bool)
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		sourceRootResolver:    SourceRootResolver{},
		importRule:            rules.ImportStatementRule{},
		relativeImportRule:    rules.RelativeImportRule{},
		importedReferenceRule: rules.ImportedModuleReferenceRule{},
		qualifiedRule:         rules.QualifiedModuleReferenceRule{},
		stringAnnotationRule:  rules.StringAnnotationRule{},
		configScanner:         DottedPathReferenceScanner{},
		commandPath:           commandPath,
	}
}

func commandPath(name string) (string, bool) {
	path, err := exec.LookPath(name)
	return path, err == nil
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (adapterproto.AggregatedResponse, bool, error) {
	_ = scanConfig

	if !plan.ContainsExtension(".py") {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	sourceRoots, err := a.sourceRootResolver.Resolve(projectRoot, plan.Moves)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	moduleMapper := NewModuleMapper(sourceRoots)
	moduleMappings, warnings := moduleMapper.Derive(plan.Moves)
	if len(moduleMappings) == 0 {
		return adapterproto.AggregatedResponse{Warnings: warnings}, true, nil
	}

	replacements, replacementWarnings, err := a.collectReplacements(projectRoot, moduleMapper, moduleMappings, scanIndex)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	configReplacements, err := a.configScanner.Scan(projectRoot, scanIndex, moduleMappings)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}
	replacements = append(replacements, configReplacements...)

	symbolMappings := make([]adapterproto.SymbolMapping, 0, len(moduleMappings))
	for _, mapping := range moduleMappings {
		symbolMappings = append(symbolMappings, mapping.ToSymbolMapping())
	}

	warnings = append(warnings, replacementWarnings...)
	return adapterproto.AggregatedResponse{
		SymbolMappings: symbolMappings,
		Replacements:   replacements,
		Warnings:       warnings,
		Checks:         pythonSanityChecks(plan, replacements, a.commandPath),
	}, true, nil
}

func (a *Analyzer) collectReplacements(projectRoot string, moduleMapper ModuleMapper, mappings []ModuleMapping, scanIndex *scan.Index) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
	pythonFiles, err := scanIndex.CandidateFiles(projectRoot, moduleCandidateQuery(mappings))
	if err != nil {
		return nil, nil, err
	}

	var allReplacements []adapterproto.Replacement
	var warnings []adapterproto.Warning
	for _, pythonFile := range pythonFiles {
		source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(pythonFile)))
		if err != nil {
			return nil, nil, err
		}
		if hasDynamicImport(string(source)) {
			warnings = append(warnings, adapterproto.Warning{
				File:    pythonFile,
				Message: "Dynamic Python import detected; not changed.",
			})
		}

		packageName, ok := moduleMapper.PackageForPath(pythonFile)
		if !ok {
			continue
		}

		document, err := Parse(source)
		if err != nil {
			warnings = append(warnings, adapterproto.Warning{
				File:    pythonFile,
				Message: "Python file not analysed because it could not be parsed",
			})
			continue
		}

		for _, mapping := range mappings {
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(a.importRule.Collect(document, rules.ImportStatementInput{
				File:      pythonFile,
				OldModule: mapping.OldModule,
				NewModule: mapping.NewModule,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(a.relativeImportRule.Collect(document, rules.RelativeImportInput{
				File:      pythonFile,
				Package:   packageName,
				OldModule: mapping.OldModule,
				NewModule: mapping.NewModule,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(a.importedReferenceRule.Collect(document, rules.ImportedModuleReferenceInput{
				File:      pythonFile,
				Package:   packageName,
				Source:    source,
				OldModule: mapping.OldModule,
				NewModule: mapping.NewModule,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(a.qualifiedRule.Collect(document, rules.QualifiedModuleReferenceInput{
				File:      pythonFile,
				OldModule: mapping.OldModule,
				NewModule: mapping.NewModule,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(a.stringAnnotationRule.Collect(document, rules.StringAnnotationInput{
				File:      pythonFile,
				OldModule: mapping.OldModule,
				NewModule: mapping.NewModule,
			}))...)
		}

		document.Close()
	}

	return allReplacements, warnings, nil
}

func moduleCandidateQuery(mappings []ModuleMapping) scan.CandidateQuery {
	query := scan.CandidateQuery{
		Extensions: []string{".py"},
	}
	for _, mapping := range mappings {
		query.IncludePaths = append(query.IncludePaths, mapping.OldPath)
		query.Needles = append(query.Needles, mapping.OldModule, mapping.OldLeaf)
	}
	return query
}

func hasDynamicImport(content string) bool {
	return strings.Contains(content, "importlib.import_module") || strings.Contains(content, "__import__(")
}

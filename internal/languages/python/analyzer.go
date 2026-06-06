//go:build cgo

package python

import (
	"os"
	"path/filepath"
	"strings"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/files"
	"refactorlah/internal/languages"
	"refactorlah/internal/languages/python/rules"
	"refactorlah/internal/planning"
)

type Analyzer struct {
	sourceRootResolver SourceRootResolver
	importRule         rules.ImportStatementRule
	relativeImportRule rules.RelativeImportRule
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		sourceRootResolver: SourceRootResolver{},
		importRule:         rules.ImportStatementRule{},
		relativeImportRule: rules.RelativeImportRule{},
	}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
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

	replacements, replacementWarnings, err := a.collectReplacements(projectRoot, moduleMapper, moduleMappings)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}

	symbolMappings := make([]adapterproto.SymbolMapping, 0, len(moduleMappings))
	for _, mapping := range moduleMappings {
		symbolMappings = append(symbolMappings, mapping.ToSymbolMapping())
	}

	warnings = append(warnings, replacementWarnings...)
	return adapterproto.AggregatedResponse{
		SymbolMappings: symbolMappings,
		Replacements:   replacements,
		Warnings:       warnings,
	}, true, nil
}

func (a *Analyzer) collectReplacements(projectRoot string, moduleMapper ModuleMapper, mappings []ModuleMapping) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
	pythonFiles, err := files.CollectFiles(projectRoot, ".")
	if err != nil {
		return nil, nil, err
	}

	var allReplacements []adapterproto.Replacement
	var warnings []adapterproto.Warning
	for _, pythonFile := range pythonFiles {
		if filepath.Ext(pythonFile) != ".py" {
			continue
		}

		source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(pythonFile)))
		if err != nil {
			return nil, nil, err
		}
		if !isPythonCandidate(pythonFile, string(source), mappings) {
			continue
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
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.importRule.Collect(document, rules.ImportStatementInput{
				File:      pythonFile,
				OldModule: mapping.OldModule,
				NewModule: mapping.NewModule,
			}))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.relativeImportRule.Collect(document, rules.RelativeImportInput{
				File:      pythonFile,
				Package:   packageName,
				OldModule: mapping.OldModule,
				NewModule: mapping.NewModule,
			}))...)
		}

		document.Close()
	}

	return allReplacements, warnings, nil
}

func isPythonCandidate(file string, content string, mappings []ModuleMapping) bool {
	for _, mapping := range mappings {
		if mapping.OldPath == file || strings.Contains(content, mapping.OldModule) || strings.Contains(content, mapping.OldLeaf) {
			return true
		}
	}
	return false
}

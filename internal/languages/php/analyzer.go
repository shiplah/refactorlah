//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"strings"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/files"
	"refactorlah/internal/languages/php/rules"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
	"refactorlah/internal/replacements"
)

type Analyzer struct {
	symbolScanner    *SymbolScanner
	namespaceRule    rules.NamespaceDeclarationRule
	classRule        rules.ClassDeclarationRule
	useStatementRule rules.UseStatementRule
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbolScanner:    NewSymbolScanner(),
		namespaceRule:    rules.NamespaceDeclarationRule{},
		classRule:        rules.ClassDeclarationRule{},
		useStatementRule: rules.UseStatementRule{},
	}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	if !plan.ContainsExtension(".php") {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	composerRoot, found, err := project.FindComposerRootForPaths(projectRoot, planPaths(plan))
	if err != nil || !found {
		return adapterproto.AggregatedResponse{}, found, err
	}

	psr4, err := ReadComposerPsr4Map(projectRoot, composerRoot)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}

	symbolMappings, warnings := a.symbolScanner.Scan(projectRoot, psr4, plan.Moves)
	replacements, replacementWarnings, err := a.collectReplacements(projectRoot, composerRoot, symbolMappings)
	if err != nil {
		return adapterproto.AggregatedResponse{}, true, err
	}

	warnings = append(warnings, replacementWarnings...)
	return adapterproto.AggregatedResponse{
		SymbolMappings: symbolMappings,
		Replacements:   replacements,
		Warnings:       warnings,
	}, true, nil
}

func (a *Analyzer) collectReplacements(projectRoot string, composerRoot string, mappings []adapterproto.SymbolMapping) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
	if len(mappings) == 0 {
		return nil, nil, nil
	}

	phpFiles, err := collectPhpFiles(projectRoot, composerRoot)
	if err != nil {
		return nil, nil, err
	}

	movedFiles := map[string]adapterproto.SymbolMapping{}
	for _, mapping := range mappings {
		movedFiles[mapping.OldPath] = mapping
	}

	var allReplacements []adapterproto.Replacement
	var warnings []adapterproto.Warning
	for _, phpFile := range phpFiles {
		source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(phpFile)))
		if err != nil {
			return nil, nil, err
		}

		document, err := Parse(source)
		if err != nil {
			warnings = append(warnings, adapterproto.Warning{
				File:    phpFile,
				Message: "PHP file not analysed because it could not be parsed",
			})
			continue
		}

		if mapping, ok := movedFiles[phpFile]; ok {
			allReplacements = append(allReplacements, convertReplacements(a.namespaceRule.Collect(document, rules.NamespaceDeclarationInput{
				File:         phpFile,
				OldNamespace: mapping.OldNamespace,
				NewNamespace: mapping.NewNamespace,
			}))...)
			allReplacements = append(allReplacements, convertReplacements(a.classRule.Collect(document, rules.ClassDeclarationInput{
				File:         phpFile,
				OldShortName: shortSymbolName(mapping.OldSymbol),
				NewShortName: shortSymbolName(mapping.NewSymbol),
			}))...)
		}

		for _, mapping := range mappings {
			allReplacements = append(allReplacements, convertReplacements(a.useStatementRule.Collect(document, rules.UseStatementInput{
				File:      phpFile,
				OldSymbol: mapping.OldSymbol,
				NewSymbol: mapping.NewSymbol,
			}))...)
		}

		document.Close()
	}

	return allReplacements, warnings, nil
}

func collectPhpFiles(projectRoot string, composerRoot string) ([]string, error) {
	collected, err := files.CollectFiles(composerRoot, ".")
	if err != nil {
		return nil, err
	}

	phpFiles := make([]string, 0, len(collected))
	for _, relativeToComposer := range collected {
		if filepath.Ext(relativeToComposer) != ".php" {
			continue
		}

		absolutePath := filepath.Join(composerRoot, filepath.FromSlash(relativeToComposer))
		projectRelative, err := filepath.Rel(projectRoot, absolutePath)
		if err != nil {
			return nil, err
		}
		phpFiles = append(phpFiles, filepath.ToSlash(projectRelative))
	}

	return phpFiles, nil
}

func convertReplacements(input []replacements.Replacement) []adapterproto.Replacement {
	output := make([]adapterproto.Replacement, 0, len(input))
	for _, replacement := range input {
		output = append(output, adapterproto.Replacement{
			File:        replacement.File,
			Start:       replacement.Start,
			End:         replacement.End,
			Replacement: replacement.Replacement,
			Reason:      replacement.Reason,
			Rule:        replacement.Rule,
			Adapter:     replacement.Adapter,
		})
	}
	return output
}

func shortSymbolName(symbol string) string {
	separator := strings.LastIndex(symbol, "\\")
	if separator < 0 {
		return symbol
	}
	return symbol[separator+1:]
}

func planPaths(plan planning.MovePlan) []string {
	paths := make([]string, 0, len(plan.Moves)*2)
	for _, move := range plan.Moves {
		paths = append(paths, move.OldPath, move.NewPath)
	}
	return paths
}

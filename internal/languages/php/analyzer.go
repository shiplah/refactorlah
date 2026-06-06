//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"strings"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/files"
	"refactorlah/internal/languages"
	"refactorlah/internal/languages/php/rules"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
)

type Analyzer struct {
	symbolScanner      *SymbolScanner
	namespaceRule      rules.NamespaceDeclarationRule
	classRule          rules.ClassDeclarationRule
	useStatementRule   rules.UseStatementRule
	fqcnRule           rules.FullyQualifiedClassNameRule
	classConstantRule  rules.ClassConstantRule
	shortNameRule      rules.ShortClassNameReferenceRule
	docblockVarRule    rules.DocblockVarRule
	docblockParamRule  rules.DocblockParamRule
	docblockReturnRule rules.DocblockReturnRule
	docblockThrowsRule rules.DocblockThrowsRule
	localImportRule    rules.NamespaceLocalDependencyImportRule
	importRemovalRule  rules.SameNamespaceImportRemovalRule
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbolScanner:      NewSymbolScanner(),
		namespaceRule:      rules.NamespaceDeclarationRule{},
		classRule:          rules.ClassDeclarationRule{},
		useStatementRule:   rules.UseStatementRule{},
		fqcnRule:           rules.FullyQualifiedClassNameRule{},
		classConstantRule:  rules.ClassConstantRule{},
		shortNameRule:      rules.ShortClassNameReferenceRule{},
		docblockVarRule:    rules.DocblockVarRule{},
		docblockParamRule:  rules.DocblockParamRule{},
		docblockReturnRule: rules.DocblockReturnRule{},
		docblockThrowsRule: rules.DocblockThrowsRule{},
		localImportRule:    rules.NamespaceLocalDependencyImportRule{},
		importRemovalRule:  rules.SameNamespaceImportRemovalRule{},
	}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	if !plan.ContainsExtension(".php") {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	composerRoot, found, err := project.FindComposerRootForPaths(projectRoot, languages.MovePaths(plan))
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
	mappingReferences := make([]rules.SymbolMappingReference, 0, len(mappings))
	for _, mapping := range mappings {
		movedFiles[mapping.OldPath] = mapping
		mappingReferences = append(mappingReferences, rules.SymbolMappingReference{
			OldSymbol: mapping.OldSymbol,
			NewSymbol: mapping.NewSymbol,
		})
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
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.namespaceRule.Collect(document, rules.NamespaceDeclarationInput{
				File:         phpFile,
				OldNamespace: mapping.OldNamespace,
				NewNamespace: mapping.NewNamespace,
			}))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.classRule.Collect(document, rules.ClassDeclarationInput{
				File:         phpFile,
				OldShortName: shortSymbolName(mapping.OldSymbol),
				NewShortName: shortSymbolName(mapping.NewSymbol),
			}))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.localImportRule.Collect(document, rules.NamespaceLocalDependencyImportInput{
				File:         phpFile,
				Source:       source,
				OldNamespace: mapping.OldNamespace,
				NewNamespace: mapping.NewNamespace,
				Mappings:     mappingReferences,
			}))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.importRemovalRule.Collect(document, rules.SameNamespaceImportRemovalInput{
				File:         phpFile,
				Source:       source,
				NewNamespace: mapping.NewNamespace,
				Mappings:     mappingReferences,
			}))...)
		}

		sameNamespaceRemovalNamespace := ""
		if movedMapping, ok := movedFiles[phpFile]; ok {
			sameNamespaceRemovalNamespace = movedMapping.NewNamespace
		}

		for _, mapping := range mappings {
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.useStatementRule.Collect(document, rules.UseStatementInput{
				File:                          phpFile,
				OldSymbol:                     mapping.OldSymbol,
				NewSymbol:                     mapping.NewSymbol,
				SameNamespaceRemovalNamespace: sameNamespaceRemovalNamespace,
			}))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.fqcnRule.Collect(document, rules.FullyQualifiedClassNameInput{
				File:      phpFile,
				OldSymbol: mapping.OldSymbol,
				NewSymbol: mapping.NewSymbol,
			}))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.classConstantRule.Collect(document, rules.ClassConstantInput{
				File:      phpFile,
				OldSymbol: mapping.OldSymbol,
				NewSymbol: mapping.NewSymbol,
			}))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.shortNameRule.Collect(document, rules.ShortClassNameReferenceInput{
				File:      phpFile,
				Source:    source,
				OldSymbol: mapping.OldSymbol,
				NewSymbol: mapping.NewSymbol,
			}))...)
			symbolInput := rules.SymbolReferenceInput{
				File:         phpFile,
				Source:       source,
				OldSymbol:    mapping.OldSymbol,
				NewSymbol:    mapping.NewSymbol,
				OldNamespace: mapping.OldNamespace,
				NewNamespace: mapping.NewNamespace,
			}
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.docblockVarRule.Collect(document, symbolInput))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.docblockParamRule.Collect(document, symbolInput))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.docblockReturnRule.Collect(document, symbolInput))...)
			allReplacements = append(allReplacements, languages.ToAdapterReplacements(a.docblockThrowsRule.Collect(document, symbolInput))...)
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

func shortSymbolName(symbol string) string {
	separator := strings.LastIndex(symbol, "\\")
	if separator < 0 {
		return symbol
	}
	return symbol[separator+1:]
}

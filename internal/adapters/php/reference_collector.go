//go:build cgo

package php

import (
	"os"
	"path/filepath"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/php/names"
	"refactorlah/internal/adapters/php/rules"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/adapters/shared"
)

type ReferenceCollector struct {
	namespaceRule           rules.NamespaceDeclarationRule
	classRule               rules.ClassDeclarationRule
	useStatementRule        rules.UseStatementRule
	fqcnRule                rules.FullyQualifiedClassNameRule
	classConstantRule       rules.ClassConstantRule
	shortNameRule           rules.ShortClassNameReferenceRule
	docblockVarRule         rules.DocblockVarRule
	docblockParamRule       rules.DocblockParamRule
	docblockReturnRule      rules.DocblockReturnRule
	docblockThrowsRule      rules.DocblockThrowsRule
	candidateSelector       CandidateFileSelector
	sameNamespaceImportRule rules.SameNamespaceReferenceImportRule
	localImportRule         rules.NamespaceLocalDependencyImportRule
	importRemovalRule       rules.SameNamespaceImportRemovalRule
}

func NewReferenceCollector() ReferenceCollector {
	return ReferenceCollector{
		namespaceRule:           rules.NamespaceDeclarationRule{},
		classRule:               rules.ClassDeclarationRule{},
		useStatementRule:        rules.UseStatementRule{},
		fqcnRule:                rules.FullyQualifiedClassNameRule{},
		classConstantRule:       rules.ClassConstantRule{},
		shortNameRule:           rules.ShortClassNameReferenceRule{},
		docblockVarRule:         rules.DocblockVarRule{},
		docblockParamRule:       rules.DocblockParamRule{},
		docblockReturnRule:      rules.DocblockReturnRule{},
		docblockThrowsRule:      rules.DocblockThrowsRule{},
		candidateSelector:       CandidateFileSelector{},
		sameNamespaceImportRule: rules.SameNamespaceReferenceImportRule{},
		localImportRule:         rules.NamespaceLocalDependencyImportRule{},
		importRemovalRule:       rules.SameNamespaceImportRemovalRule{},
	}
}

func (c ReferenceCollector) Collect(projectRoot string, composerRoot string, mappings []adapterproto.SymbolMapping, scanIndex *scan.Index) ([]adapterproto.Replacement, []adapterproto.Warning, error) {
	mappingSet := NewSymbolMappingSet(mappings)
	if mappingSet.Len() == 0 {
		return nil, nil, nil
	}

	phpFiles, err := scanIndex.CandidateFiles(composerRoot, c.candidateSelector.Query(mappingSet.All()))
	if err != nil {
		return nil, nil, err
	}

	mappingReferences := mappingSet.References()
	allMappings := mappingSet.All()
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
		warnings = append(warnings, collectReferenceWarnings(document, phpFile, source, mappingReferences)...)

		if mapping, ok := mappingSet.MovedFile(phpFile); ok {
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.namespaceRule.Collect(document, rules.NamespaceDeclarationInput{
				File:         phpFile,
				OldNamespace: mapping.OldNamespace,
				NewNamespace: mapping.NewNamespace,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.classRule.Collect(document, rules.ClassDeclarationInput{
				File:         phpFile,
				OldShortName: names.Short(mapping.OldSymbol),
				NewShortName: names.Short(mapping.NewSymbol),
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.localImportRule.Collect(document, rules.NamespaceLocalDependencyImportInput{
				File:         phpFile,
				Source:       source,
				OldNamespace: mapping.OldNamespace,
				NewNamespace: mapping.NewNamespace,
				Mappings:     mappingReferences,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.importRemovalRule.Collect(document, rules.SameNamespaceImportRemovalInput{
				File:         phpFile,
				Source:       source,
				NewNamespace: mapping.NewNamespace,
				Mappings:     mappingReferences,
			}))...)
		}

		sameNamespaceRemovalNamespace := ""
		if movedMapping, ok := mappingSet.MovedFile(phpFile); ok {
			sameNamespaceRemovalNamespace = movedMapping.NewNamespace
		} else {
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.sameNamespaceImportRule.Collect(document, rules.SameNamespaceReferenceImportInput{
				File:     phpFile,
				Source:   source,
				Mappings: mappingReferences,
			}))...)
		}

		for _, mapping := range allMappings {
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.useStatementRule.Collect(document, rules.UseStatementInput{
				File:                          phpFile,
				OldSymbol:                     mapping.OldSymbol,
				NewSymbol:                     mapping.NewSymbol,
				SameNamespaceRemovalNamespace: sameNamespaceRemovalNamespace,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.fqcnRule.Collect(document, rules.FullyQualifiedClassNameInput{
				File:      phpFile,
				OldSymbol: mapping.OldSymbol,
				NewSymbol: mapping.NewSymbol,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.classConstantRule.Collect(document, rules.ClassConstantInput{
				File:      phpFile,
				OldSymbol: mapping.OldSymbol,
				NewSymbol: mapping.NewSymbol,
			}))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.shortNameRule.Collect(document, rules.ShortClassNameReferenceInput{
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
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.docblockVarRule.Collect(document, symbolInput))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.docblockParamRule.Collect(document, symbolInput))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.docblockReturnRule.Collect(document, symbolInput))...)
			allReplacements = append(allReplacements, shared.ToAdapterReplacements(c.docblockThrowsRule.Collect(document, symbolInput))...)
		}

		document.Close()
	}

	return allReplacements, warnings, nil
}

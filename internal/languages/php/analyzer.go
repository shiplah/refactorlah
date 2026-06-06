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
	"refactorlah/internal/languages/php/symfony/core"
	"refactorlah/internal/languages/php/symfony/twig"
	"refactorlah/internal/languages/staticimports"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
)

type Analyzer struct {
	symbolScanner             *SymbolScanner
	namespaceRule             rules.NamespaceDeclarationRule
	classRule                 rules.ClassDeclarationRule
	useStatementRule          rules.UseStatementRule
	fqcnRule                  rules.FullyQualifiedClassNameRule
	classConstantRule         rules.ClassConstantRule
	shortNameRule             rules.ShortClassNameReferenceRule
	docblockVarRule           rules.DocblockVarRule
	docblockParamRule         rules.DocblockParamRule
	docblockReturnRule        rules.DocblockReturnRule
	docblockThrowsRule        rules.DocblockThrowsRule
	localImportRule           rules.NamespaceLocalDependencyImportRule
	importRemovalRule         rules.SameNamespaceImportRemovalRule
	semanticHintScanner       SemanticHintScanner
	twigConfigReader          twig.ConfigReader
	twigMapper                twig.TemplateMapper
	twigRuleRegistry          twig.RuleRegistry
	componentNamespaceScanner twig.ComponentNamespaceScanner
	staticImportScanner       staticimports.Scanner
	assetMapperScanner        core.AssetMapperScanner
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbolScanner:             NewSymbolScanner(),
		namespaceRule:             rules.NamespaceDeclarationRule{},
		classRule:                 rules.ClassDeclarationRule{},
		useStatementRule:          rules.UseStatementRule{},
		fqcnRule:                  rules.FullyQualifiedClassNameRule{},
		classConstantRule:         rules.ClassConstantRule{},
		shortNameRule:             rules.ShortClassNameReferenceRule{},
		docblockVarRule:           rules.DocblockVarRule{},
		docblockParamRule:         rules.DocblockParamRule{},
		docblockReturnRule:        rules.DocblockReturnRule{},
		docblockThrowsRule:        rules.DocblockThrowsRule{},
		localImportRule:           rules.NamespaceLocalDependencyImportRule{},
		importRemovalRule:         rules.SameNamespaceImportRemovalRule{},
		semanticHintScanner:       SemanticHintScanner{},
		twigConfigReader:          twig.ConfigReader{},
		twigMapper:                twig.TemplateMapper{},
		twigRuleRegistry:          twig.NewRuleRegistry(),
		componentNamespaceScanner: twig.ComponentNamespaceScanner{},
		staticImportScanner:       staticimports.Scanner{},
		assetMapperScanner:        core.AssetMapperScanner{},
	}
}

func (a *Analyzer) Analyze(projectRoot string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	containsPHP := plan.ContainsExtension(".php")
	containsTwig := plan.ContainsExtension(".twig")
	containsStaticImport := planContainsStaticImportTarget(plan)
	containsProjectDirectoryPath := plan.IsDir
	if !containsPHP && !containsTwig && !containsStaticImport && !containsProjectDirectoryPath {
		return adapterproto.AggregatedResponse{}, false, nil
	}

	composerRoot, found, err := project.FindComposerRootForPaths(projectRoot, languages.MovePaths(plan))
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
		phpReplacements, replacementWarnings, err := a.collectReplacements(projectRoot, composerRoot, phpSymbolMappings)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}
		yamlReplacements, err := a.collectYamlSymbolReplacements(projectRoot, composerRoot, phpSymbolMappings)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}
		semanticWarnings, err := a.collectSemanticWarnings(projectRoot, composerRoot, phpSymbolMappings)
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
		twigPathMappings, twigReplacements, twigWarnings, err := a.collectTwig(projectRoot, composerRoot, plan)
		if err != nil {
			return adapterproto.AggregatedResponse{}, true, err
		}
		pathMappings = append(pathMappings, twigPathMappings...)
		replacements = append(replacements, twigReplacements...)
		warnings = append(warnings, twigWarnings...)
	}

	projectPathMappings, pathReplacements, err := a.collectProjectPathReplacements(projectRoot, composerRoot, plan, containsStaticImport)
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

func (a *Analyzer) collectYamlSymbolReplacements(projectRoot string, composerRoot string, mappings []adapterproto.SymbolMapping) ([]adapterproto.Replacement, error) {
	if len(mappings) == 0 {
		return nil, nil
	}

	yamlFiles, err := collectFilesByExtension(projectRoot, composerRoot, ".yaml", ".yml")
	if err != nil {
		return nil, err
	}
	componentNamespaceReplacements, err := a.componentNamespaceScanner.Scan(projectRoot, yamlFiles, mappings)
	if err != nil {
		return nil, err
	}

	return languages.ToAdapterReplacements(componentNamespaceReplacements), nil
}

func (a *Analyzer) collectSemanticWarnings(projectRoot string, composerRoot string, mappings []adapterproto.SymbolMapping) ([]adapterproto.Warning, error) {
	if len(mappings) == 0 {
		return nil, nil
	}

	phpFiles, err := collectPhpFiles(projectRoot, composerRoot)
	if err != nil {
		return nil, err
	}
	textFiles, err := collectFilesByExtension(projectRoot, composerRoot, ".yaml", ".yml", ".xml", ".neon")
	if err != nil {
		return nil, err
	}

	return a.semanticHintScanner.Scan(projectRoot, phpFiles, textFiles, mappings)
}

func (a *Analyzer) collectProjectPathReplacements(projectRoot string, composerRoot string, plan planning.MovePlan, containsStaticImport bool) ([]adapterproto.PathMapping, []adapterproto.Replacement, error) {
	var allReplacements []adapterproto.Replacement

	if containsStaticImport {
		staticFiles, err := collectStaticImportFiles(projectRoot, composerRoot)
		if err != nil {
			return nil, nil, err
		}
		staticReplacements, err := a.staticImportScanner.Scan(projectRoot, staticFiles, plan.Moves)
		if err != nil {
			return nil, nil, err
		}
		allReplacements = append(allReplacements, languages.ToAdapterReplacements(staticReplacements)...)
	}

	projectPathMappings := core.ProjectDirectoryPathMappings(plan)
	if len(projectPathMappings) == 0 {
		return nil, allReplacements, nil
	}

	yamlFiles, err := collectFilesByExtension(projectRoot, composerRoot, ".yaml", ".yml")
	if err != nil {
		return nil, nil, err
	}
	assetMapperReplacements, err := a.assetMapperScanner.Scan(projectRoot, yamlFiles, projectPathMappings)
	if err != nil {
		return nil, nil, err
	}
	allReplacements = append(allReplacements, languages.ToAdapterReplacements(assetMapperReplacements)...)

	return projectPathMappings, allReplacements, nil
}

func (a *Analyzer) collectTwig(projectRoot string, composerRoot string, plan planning.MovePlan) ([]adapterproto.PathMapping, []adapterproto.Replacement, []adapterproto.Warning, error) {
	configuration, err := a.twigConfigReader.ReadFromConfigRoot(projectRoot, composerRoot)
	if err != nil {
		return nil, nil, nil, err
	}

	pathMappings := a.twigMapper.DeriveMappings(plan.Moves, configuration)
	if len(pathMappings) == 0 {
		return nil, nil, nil, nil
	}

	phpConfigFiles, twigFiles, err := collectTwigReferenceFiles(projectRoot, composerRoot)
	if err != nil {
		return nil, nil, nil, err
	}

	twigReplacements, warnings, err := a.twigRuleRegistry.Scan(projectRoot, phpConfigFiles, twigFiles, pathMappings)
	if err != nil {
		return nil, nil, nil, err
	}

	return pathMappings, languages.ToAdapterReplacements(twigReplacements), warnings, nil
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

func collectTwigReferenceFiles(projectRoot string, composerRoot string) ([]string, []string, error) {
	collected, err := files.CollectFiles(composerRoot, ".")
	if err != nil {
		return nil, nil, err
	}

	var filesToScan []string
	var twigFiles []string
	for _, relativeToComposer := range collected {
		extension := filepath.Ext(relativeToComposer)
		if extension != ".php" && extension != ".yaml" && extension != ".yml" && extension != ".twig" {
			continue
		}

		projectRelativeSlash, err := composerRelativeToProjectRelative(projectRoot, composerRoot, relativeToComposer)
		if err != nil {
			return nil, nil, err
		}
		if extension == ".twig" {
			twigFiles = append(twigFiles, projectRelativeSlash)
			continue
		}
		filesToScan = append(filesToScan, projectRelativeSlash)
	}

	return filesToScan, twigFiles, nil
}

func collectStaticImportFiles(projectRoot string, composerRoot string) ([]string, error) {
	return collectFilesByExtension(projectRoot, composerRoot, ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".css")
}

func collectFilesByExtension(projectRoot string, composerRoot string, extensions ...string) ([]string, error) {
	collected, err := files.CollectFiles(composerRoot, ".")
	if err != nil {
		return nil, err
	}

	wanted := map[string]bool{}
	for _, extension := range extensions {
		wanted[extension] = true
	}

	var result []string
	for _, relativeToComposer := range collected {
		extension := filepath.Ext(relativeToComposer)
		if !wanted[extension] {
			continue
		}

		projectRelativeSlash, err := composerRelativeToProjectRelative(projectRoot, composerRoot, relativeToComposer)
		if err != nil {
			return nil, err
		}
		result = append(result, projectRelativeSlash)
	}

	return result, nil
}

func composerRelativeToProjectRelative(projectRoot string, composerRoot string, relativeToComposer string) (string, error) {
	absolutePath := filepath.Join(composerRoot, filepath.FromSlash(relativeToComposer))
	projectRelative, err := filepath.Rel(projectRoot, absolutePath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(projectRelative), nil
}

func planContainsStaticImportTarget(plan planning.MovePlan) bool {
	for _, extension := range []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".css"} {
		if plan.ContainsExtension(extension) {
			return true
		}
	}
	return false
}

func shortSymbolName(symbol string) string {
	separator := strings.LastIndex(symbol, "\\")
	if separator < 0 {
		return symbol
	}
	return symbol[separator+1:]
}

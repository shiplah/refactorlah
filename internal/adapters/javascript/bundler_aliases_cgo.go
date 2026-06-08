//go:build cgo

package javascript

import (
	"os"
	"path/filepath"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

type bundlerAliasConfig struct {
	fileNames []string
	reason    string
	rule      string
	label     string
}

var bundlerAliasConfigs = []bundlerAliasConfig{
	{
		fileNames: []string{"vite.config.js", "vite.config.ts", "vite.config.mjs", "vite.config.cjs"},
		reason:    rules.ViteAliasReason,
		rule:      rules.ViteAliasRuleName,
		label:     "Vite",
	},
	{
		fileNames: []string{"webpack.config.js", "webpack.config.ts", "webpack.config.mjs", "webpack.config.cjs"},
		reason:    rules.WebpackAliasReason,
		rule:      rules.WebpackAliasRuleName,
		label:     "webpack",
	},
}

func (a *Analyzer) collectBundlerAliasReplacements(projectRoot string, plan planning.MovePlan, scanIndex *scan.Index) ([]replacements.Replacement, []adapterproto.Warning, error) {
	var allReplacements []replacements.Replacement
	var warnings []adapterproto.Warning
	for _, config := range bundlerAliasConfigs {
		for _, file := range config.fileNames {
			mappings, configWarnings, err := readBundlerAliasMappings(projectRoot, file, config)
			if err != nil {
				return nil, nil, err
			}
			warnings = append(warnings, configWarnings...)
			rewrites := rules.PathAliasSpecifierRule{Reason: config.reason, Rule: config.rule}.Collect(mappings, plan.Moves)
			if len(rewrites) == 0 {
				continue
			}

			files, err := scanIndex.CandidateFiles(projectRoot, rules.SpecifierRewriteCandidateQuery(rewrites))
			if err != nil {
				return nil, nil, err
			}
			configReplacements, err := a.scanner.ScanSpecifiers(projectRoot, files, rewrites)
			if err != nil {
				return nil, nil, err
			}
			allReplacements = append(allReplacements, configReplacements...)
		}
	}
	return allReplacements, warnings, nil
}

func readBundlerAliasMappings(projectRoot string, file string, config bundlerAliasConfig) ([]rules.PathAliasMapping, []adapterproto.Warning, error) {
	configPath := filepath.Join(projectRoot, file)
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	document, err := parseJavaScriptConfig(file, content)
	if err != nil {
		return nil, []adapterproto.Warning{{
			File:    file,
			Message: config.label + " config could not be parsed; alias rewrites skipped.",
		}}, nil
	}
	defer document.Close()

	mappings := bundlerAliasMappingsFromDocument(projectRoot, document.RootNode(), config)
	return mappings, nil, nil
}

func bundlerAliasMapping(projectRoot string, alias string, replacement string) (rules.PathAliasMapping, bool) {
	if alias == "" || strings.Contains(alias, "*") || !filepath.IsAbs(replacement) {
		return rules.PathAliasMapping{}, false
	}

	relative, err := filepath.Rel(projectRoot, replacement)
	if err != nil {
		return rules.PathAliasMapping{}, false
	}
	relative = filepath.ToSlash(relative)
	if relative == ".." || filepath.IsAbs(relative) || rules.StartsWithParentTraversal(relative) {
		return rules.PathAliasMapping{}, false
	}

	aliasPrefix := alias
	if !strings.HasSuffix(aliasPrefix, "/") {
		aliasPrefix += "/"
	}
	targetPrefix := ""
	if relative != "." {
		targetPrefix = strings.TrimSuffix(relative, "/") + "/"
	}
	return rules.PathAliasMapping{
		AliasPrefix:  aliasPrefix,
		TargetPrefix: targetPrefix,
	}, true
}

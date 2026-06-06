package twig

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	adapterproto "refactorlah/internal/adapters"
	twigrules "refactorlah/internal/languages/php/symfony/twig/rules"
	"refactorlah/internal/replacements"
)

type RuleRegistry struct {
	twigRules []twigrules.StringReplacementRule
	phpRules  []twigrules.StringReplacementRule
	yamlRules []twigrules.StringReplacementRule
}

func NewRuleRegistry() RuleRegistry {
	return RuleRegistry{
		twigRules: []twigrules.StringReplacementRule{
			twigrules.TwigIncludeRule(),
			twigrules.TwigExtendsRule(),
			twigrules.TwigEmbedRule(),
			twigrules.TwigUseRule(),
			twigrules.TwigImportRule(),
			twigrules.TwigFromRule(),
		},
		phpRules: []twigrules.StringReplacementRule{
			twigrules.RenderTemplateRule(),
			twigrules.TemplateAttributeRule(),
			twigrules.ComponentTemplateAttributeRule(),
		},
		yamlRules: []twigrules.StringReplacementRule{
			twigrules.YamlTemplateRule(),
			twigrules.YamlComponentTemplateDirectoryRule(),
		},
	}
}

func (r RuleRegistry) Scan(projectRoot string, files []string, twigFiles []string, mappings []adapterproto.PathMapping) ([]replacements.Replacement, []adapterproto.Warning, error) {
	if len(mappings) == 0 {
		return nil, nil, nil
	}

	var allReplacements []replacements.Replacement
	var warnings []adapterproto.Warning
	for _, file := range twigFiles {
		content, err := readProjectFile(projectRoot, file)
		if err != nil {
			return nil, nil, err
		}
		if !containsOldPathReference(content, mappings) {
			continue
		}

		allReplacements = append(allReplacements, collectPathReplacements(file, content, mappings, r.twigRules)...)
		warnings = append(warnings, twigWarnings(file, content, mappings)...)
	}

	for _, file := range files {
		content, err := readProjectFile(projectRoot, file)
		if err != nil {
			return nil, nil, err
		}
		if !containsOldPathReference(content, mappings) {
			continue
		}

		if strings.HasSuffix(file, ".php") {
			allReplacements = append(allReplacements, collectPathReplacements(file, content, mappings, r.phpRules)...)
			warnings = append(warnings, phpWarnings(file, content, mappings)...)
			continue
		}

		allReplacements = append(allReplacements, collectPathReplacements(file, content, mappings, r.yamlRules)...)
	}

	return allReplacements, warnings, nil
}

func collectPathReplacements(file string, content string, mappings []adapterproto.PathMapping, rules []twigrules.StringReplacementRule) []replacements.Replacement {
	var result []replacements.Replacement
	for _, mapping := range mappings {
		for _, rule := range rules {
			result = append(result, rule.Collect(file, content, mapping)...)
		}
	}
	return result
}

func readProjectFile(projectRoot string, file string) (string, error) {
	content, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func containsOldPathReference(content string, mappings []adapterproto.PathMapping) bool {
	for _, mapping := range mappings {
		if mapping.OldReference != "" && strings.Contains(content, mapping.OldReference) {
			return true
		}
	}
	return false
}

func twigWarnings(file string, content string, mappings []adapterproto.PathMapping) []adapterproto.Warning {
	patterns := []string{
		`{%\s*include\s+([A-Za-z_][^%\s]*)`,
		`\{\{\s*include\(\s*([A-Za-z_][^)]+)\)`,
		`{%\s*extends\s+([A-Za-z_][^%\s]*)`,
	}

	var warnings []adapterproto.Warning
	for _, pattern := range patterns {
		warnings = append(warnings, dynamicTemplateWarnings(file, content, pattern, mappings)...)
	}
	return warnings
}

func phpWarnings(file string, content string, mappings []adapterproto.PathMapping) []adapterproto.Warning {
	return dynamicTemplateWarnings(file, content, `->render(?:View)?\(\s*([^'"][^,\)]*)`, mappings)
}

func dynamicTemplateWarnings(file string, content string, pattern string, mappings []adapterproto.PathMapping) []adapterproto.Warning {
	expression := regexp.MustCompile(pattern)
	var warnings []adapterproto.Warning
	for _, match := range expression.FindAllStringSubmatchIndex(content, -1) {
		if len(match) < 4 || match[2] < 0 {
			continue
		}
		value := strings.TrimSpace(content[match[2]:match[3]])
		if value == "" || !containsWarningIndicator(value, mappings) {
			continue
		}

		warnings = append(warnings, adapterproto.Warning{
			File:    file,
			Line:    strings.Count(content[:match[2]], "\n") + 1,
			Message: "Dynamic Twig template path detected; not changed.",
		})
	}
	return warnings
}

func containsWarningIndicator(value string, mappings []adapterproto.PathMapping) bool {
	for _, mapping := range mappings {
		for _, indicator := range []string{mapping.OldReference, path.Base(mapping.OldReference)} {
			if indicator != "" && strings.Contains(value, indicator) {
				return true
			}
		}
	}
	return false
}

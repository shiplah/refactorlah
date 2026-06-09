package twig

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/replacements"
)

type ComponentNamespaceScanner struct{}

func (s ComponentNamespaceScanner) Scan(projectRoot string, files []string, mappings []adapterproto.SymbolMapping) ([]replacements.Replacement, error) {
	if len(mappings) == 0 {
		return nil, nil
	}

	var result []replacements.Replacement
	for _, file := range files {
		contentBytes, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		content := string(contentBytes)
		if !strings.Contains(content, "twig_component") {
			continue
		}

		for _, mapping := range mappings {
			result = append(result, componentNamespaceReplacements(file, content, mapping)...)
		}
	}

	return result, nil
}

func componentNamespaceReplacements(file string, content string, mapping adapterproto.SymbolMapping) []replacements.Replacement {
	if mapping.OldNamespace == "" || mapping.OldNamespace == mapping.NewNamespace {
		return nil
	}

	oldReference := mapping.OldNamespace + "\\"
	newReference := mapping.NewNamespace + "\\"
	pattern := regexp.MustCompile(`(['"])` + regexp.QuoteMeta(oldReference) + `(['"])\s*:`)

	var result []replacements.Replacement
	for _, match := range pattern.FindAllStringSubmatchIndex(content, -1) {
		if len(match) < 6 || match[2] < 0 || match[3] < match[2] || match[4] < 0 || match[5] < match[4] {
			continue
		}
		if content[match[2]:match[3]] != content[match[4]:match[5]] {
			continue
		}

		quote := content[match[2]:match[3]]
		result = append(result, replacements.Replacement{
			File:        file,
			Start:       match[2],
			End:         match[5],
			Replacement: quote + newReference + quote,
			Reason:      "yaml-twig-component-namespace",
			Rule:        "php.symfony.twig.ComponentNamespaceScanner",
			Adapter:     "php",
		})
	}

	return result
}

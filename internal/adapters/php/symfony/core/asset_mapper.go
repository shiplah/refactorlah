package core

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/replacements"
)

type AssetMapperScanner struct{}

func (s AssetMapperScanner) Scan(projectRoot string, files []string, mappings []adapterproto.PathMapping) ([]replacements.Replacement, error) {
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
		if !strings.Contains(content, "asset_mapper") {
			continue
		}

		for _, mapping := range mappings {
			result = append(result, assetMapperPathReplacements(file, content, mapping)...)
		}
	}

	return result, nil
}

func assetMapperPathReplacements(file string, content string, mapping adapterproto.PathMapping) []replacements.Replacement {
	if mapping.Kind != "project-path-directory" {
		return nil
	}

	pattern := regexp.MustCompile(`(?m)^(\s*-\s*)(['"])` + regexp.QuoteMeta(mapping.OldReference) + `(['"])\s*$`)
	var result []replacements.Replacement
	for _, match := range pattern.FindAllStringSubmatchIndex(content, -1) {
		if len(match) < 8 || match[4] < 0 || match[5] < match[4] {
			continue
		}
		if content[match[4]:match[5]] != content[match[6]:match[7]] {
			continue
		}

		start := match[4]
		end := match[7]
		quote := content[match[4]:match[5]]
		result = append(result, replacements.Replacement{
			File:        file,
			Start:       start,
			End:         end,
			Replacement: quote + mapping.NewReference + quote,
			Reason:      "yaml-asset-mapper-path",
			Rule:        "php.symfony.core.AssetMapperScanner",
			Adapter:     "php",
		})
	}

	return result
}

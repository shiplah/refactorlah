package python

import (
	"os"
	"path/filepath"
	"strings"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
)

const DottedPathReferenceRuleName = "python.DottedPathReferenceScanner"

var configReferenceExtensions = map[string]bool{
	".cfg":  true,
	".ini":  true,
	".toml": true,
	".yaml": true,
	".yml":  true,
}

type DottedPathReferenceScanner struct{}

func (s DottedPathReferenceScanner) Scan(projectRoot string, scanIndex *scan.Index, mappings []ModuleMapping) ([]adapterproto.Replacement, error) {
	if len(mappings) == 0 {
		return nil, nil
	}

	collected, err := scanIndex.CandidateFiles(projectRoot, configReferenceCandidateQuery(mappings))
	if err != nil {
		return nil, err
	}

	var replacements []adapterproto.Replacement
	for _, file := range collected {
		if !configReferenceExtensions[filepath.Ext(file)] {
			continue
		}

		contentBytes, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		content := string(contentBytes)

		for _, mapping := range mappings {
			if !strings.Contains(content, mapping.OldModule) {
				continue
			}
			replacements = append(replacements, dottedPathReplacements(file, content, mapping)...)
		}
	}

	return replacements, nil
}

func configReferenceCandidateQuery(mappings []ModuleMapping) scan.CandidateQuery {
	query := scan.CandidateQuery{
		Extensions: []string{".cfg", ".ini", ".toml", ".yaml", ".yml"},
	}
	for _, mapping := range mappings {
		query.Needles = append(query.Needles, mapping.OldModule)
	}
	return query
}

func dottedPathReplacements(file string, content string, mapping ModuleMapping) []adapterproto.Replacement {
	var replacements []adapterproto.Replacement
	offset := 0
	for {
		index := strings.Index(content[offset:], mapping.OldModule)
		if index < 0 {
			return replacements
		}

		start := offset + index
		end := start + len(mapping.OldModule)
		if isSafeDottedPathMatch(content, start, end) && !isCommentLine(content, start) {
			replacements = append(replacements, adapterproto.Replacement{
				File:        file,
				Start:       start,
				End:         end,
				Replacement: mapping.NewModule,
				Reason:      "python-config-dotted-path",
				Rule:        DottedPathReferenceRuleName,
				Adapter:     "python",
			})
		}

		offset = end
	}
}

func isSafeDottedPathMatch(content string, start int, end int) bool {
	if start > 0 && isDotPathCharacter(content[start-1]) {
		return false
	}
	if end < len(content) && isModuleSuffixCharacter(content[end]) {
		return false
	}
	return true
}

func isDotPathCharacter(value byte) bool {
	return value == '_' || value == '.' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func isModuleSuffixCharacter(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func isCommentLine(content string, offset int) bool {
	lineStart := strings.LastIndex(content[:offset], "\n") + 1
	return strings.HasPrefix(strings.TrimLeft(content[lineStart:offset], " \t"), "#") ||
		strings.HasPrefix(strings.TrimLeft(content[lineStart:offset], " \t"), ";")
}

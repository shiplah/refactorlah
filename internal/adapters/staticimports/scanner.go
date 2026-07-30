package staticimports

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/replacements"
)

type Scanner struct{}

type SpecifierRewrite struct {
	OldSpecifier string
	NewSpecifier string
	Reason       string
	Rule         string
	Adapter      string
}

func CandidateNeedles(moves []planning.FileMove) []string {
	seen := map[string]bool{}
	var needles []string
	for _, move := range moves {
		for _, needle := range []string{move.OldPath, path.Base(move.OldPath)} {
			if needle == "" || needle == "." || seen[needle] {
				continue
			}
			seen[needle] = true
			needles = append(needles, needle)
		}
	}
	return needles
}

func (s Scanner) Scan(projectRoot string, files []string, moves []planning.FileMove) ([]replacements.Replacement, error) {
	return s.scan(projectRoot, files, moves, staticSpecifierPairs)
}

func (s Scanner) ScanSpecifiers(projectRoot string, files []string, rewrites []SpecifierRewrite) ([]replacements.Replacement, error) {
	var result []replacements.Replacement
	for _, file := range files {
		contentBytes, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		content := string(contentBytes)
		if content == "" {
			continue
		}

		for _, rewrite := range rewrites {
			if rewrite.OldSpecifier == "" || rewrite.NewSpecifier == "" || rewrite.OldSpecifier == rewrite.NewSpecifier {
				continue
			}
			if !strings.Contains(content, rewrite.OldSpecifier) {
				continue
			}
			result = append(result, replacementsForSpecifier(file, content, rewrite)...)
		}
	}

	return result, nil
}

func (s Scanner) scan(projectRoot string, files []string, moves []planning.FileMove, pairBuilder func(string, planning.FileMove) []specifierPair) ([]replacements.Replacement, error) {
	var result []replacements.Replacement
	for _, file := range files {
		contentBytes, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		content := string(contentBytes)
		if content == "" {
			continue
		}

		for _, move := range moves {
			for _, pair := range pairBuilder(file, move) {
				if !strings.Contains(content, pair.oldSpecifier) {
					continue
				}
				result = append(result, replacementsForSpecifier(file, content, SpecifierRewrite{
					OldSpecifier: pair.oldSpecifier,
					NewSpecifier: pair.newSpecifier,
					Reason:       "static-import-path",
					Rule:         "staticimports.Scanner",
				})...)
			}
		}
	}

	return result, nil
}

type specifierPair struct {
	oldSpecifier string
	newSpecifier string
}

func staticSpecifierPairs(importingFile string, move planning.FileMove) []specifierPair {
	oldSpecifier := relativeSpecifier(importingFile, move.OldPath)
	newSpecifier := relativeSpecifier(importingFile, move.NewPath)
	if oldSpecifier == newSpecifier {
		return nil
	}

	pairs := []specifierPair{{oldSpecifier: oldSpecifier, newSpecifier: newSpecifier}}
	if strings.HasPrefix(oldSpecifier, "./") && strings.HasPrefix(newSpecifier, "./") {
		pairs = append(pairs, specifierPair{
			oldSpecifier: strings.TrimPrefix(oldSpecifier, "./"),
			newSpecifier: strings.TrimPrefix(newSpecifier, "./"),
		})
	}
	return pairs
}

func relativeSpecifier(importingFile string, targetPath string) string {
	fromParts := pathParts(path.Dir(importingFile))
	targetParts := pathParts(targetPath)

	common := 0
	for common < len(fromParts) && common < len(targetParts) && fromParts[common] == targetParts[common] {
		common++
	}

	relativeParts := append(make([]string, 0, len(fromParts)-common+len(targetParts)-common), stringsRepeat("..", len(fromParts)-common)...)
	relativeParts = append(relativeParts, targetParts[common:]...)
	relative := strings.Join(relativeParts, "/")
	if relative == "" {
		return "./" + lastPart(targetPath)
	}
	if !strings.HasPrefix(relative, "..") {
		return "./" + relative
	}
	return relative
}

func pathParts(input string) []string {
	trimmed := strings.Trim(input, "/.")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func lastPart(input string) string {
	parts := pathParts(input)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func stringsRepeat(value string, count int) []string {
	result := make([]string, count)
	for index := range result {
		result[index] = value
	}
	return result
}

func replacementsForSpecifier(file string, content string, rewrite SpecifierRewrite) []replacements.Replacement {
	var result []replacements.Replacement
	seenRanges := map[string]bool{}
	for _, quote := range []string{"'", `"`} {
		quotedSpecifier := regexp.QuoteMeta(quote + rewrite.OldSpecifier + quote)
		patterns := []*regexp.Regexp{
			regexp.MustCompile(`\bimport\s+(?:[^;'"]+\s+from\s*)?(` + quotedSpecifier + `)`),
			regexp.MustCompile(`\bexport\s+[^;'"]+\s+from\s*(` + quotedSpecifier + `)`),
			regexp.MustCompile(`\bimport\s*\(\s*(` + quotedSpecifier + `)\s*\)`),
			regexp.MustCompile(`\brequire\s*\(\s*(` + quotedSpecifier + `)\s*\)`),
			regexp.MustCompile(`@import\s+(?:url\(\s*)?(` + quotedSpecifier + `)`),
		}

		for _, pattern := range patterns {
			for _, match := range pattern.FindAllStringSubmatchIndex(content, -1) {
				if len(match) < 4 || match[2] < 0 {
					continue
				}
				start := match[2] + 1
				end := match[3] - 1
				rangeKey := file + ":" + strconv.Itoa(start) + ":" + strconv.Itoa(end)
				if seenRanges[rangeKey] {
					continue
				}
				seenRanges[rangeKey] = true
				result = append(result, replacements.Replacement{
					File:        file,
					Start:       start,
					End:         end,
					Replacement: rewrite.NewSpecifier,
					Reason:      replacementReason(rewrite),
					Rule:        replacementRule(rewrite),
					Adapter:     rewrite.Adapter,
				})
			}
		}
	}
	return result
}

func replacementReason(rewrite SpecifierRewrite) string {
	if rewrite.Reason != "" {
		return rewrite.Reason
	}
	return "static-import-path"
}

func replacementRule(rewrite SpecifierRewrite) string {
	if rewrite.Rule != "" {
		return rewrite.Rule
	}
	return "staticimports.Scanner"
}

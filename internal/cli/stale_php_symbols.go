package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/files"
	"github.com/shiplah/refactorlah/internal/reporting"
	"github.com/shiplah/refactorlah/internal/validation"
)

type stalePHPSymbolHit struct {
	File   string
	Line   int
	Symbol string
}

func stalePHPSymbolValidation(projectRoot string, scanConfig config.Config, mappings []adapterproto.SymbolMapping) (reporting.ValidationResult, error) {
	result := reporting.ValidationResult{Name: "stale PHP symbol scan"}
	symbols := movedPHPSymbols(mappings)
	if len(symbols) == 0 {
		return reporting.ValidationResult{}, nil
	}

	hits, err := scanStalePHPSymbols(projectRoot, scanConfig, symbols)
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
		return result, err
	}
	if len(hits) == 0 {
		result.Status = "ok"
		result.Message = "passed"
		return result, nil
	}

	result.Status = "failed"
	result.Message = fmt.Sprintf("%d stale PHP symbol reference(s) found", len(hits))
	result.Stdout = renderStalePHPSymbolHits(hits)
	return result, validation.ErrValidationFailed
}

func movedPHPSymbols(mappings []adapterproto.SymbolMapping) []string {
	index := map[string]struct{}{}
	for _, mapping := range mappings {
		if !isPHPSymbolKind(mapping.Kind) || mapping.OldSymbol == "" || mapping.OldSymbol == mapping.NewSymbol {
			continue
		}
		index[mapping.OldSymbol] = struct{}{}
	}

	symbols := make([]string, 0, len(index))
	for symbol := range index {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	return symbols
}

func isPHPSymbolKind(kind string) bool {
	switch kind {
	case "class", "interface", "trait", "enum":
		return true
	default:
		return false
	}
}

func scanStalePHPSymbols(projectRoot string, scanConfig config.Config, symbols []string) ([]stalePHPSymbolHit, error) {
	collected, err := files.CollectFiles(projectRoot, ".")
	if err != nil {
		return nil, err
	}

	var hits []stalePHPSymbolHit
	for _, file := range collected {
		if filepath.Ext(file) != ".php" || !scanConfig.Allows(file) {
			continue
		}

		contentBytes, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		content := string(contentBytes)
		for _, symbol := range symbols {
			for _, offset := range codeSymbolOffsets(content, symbol) {
				hits = append(hits, stalePHPSymbolHit{
					File:   file,
					Line:   lineForOffset(content, offset),
					Symbol: symbol,
				})
			}
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].File == hits[j].File {
			if hits[i].Line == hits[j].Line {
				return hits[i].Symbol < hits[j].Symbol
			}
			return hits[i].Line < hits[j].Line
		}
		return hits[i].File < hits[j].File
	})
	return hits, nil
}

func codeSymbolOffsets(content string, symbol string) []int {
	const (
		stateCode = iota
		stateSingleQuote
		stateDoubleQuote
		stateLineComment
		stateBlockComment
	)

	state := stateCode
	escaped := false
	var offsets []int
	for index := 0; index < len(content); {
		switch state {
		case stateSingleQuote:
			if escaped {
				escaped = false
			} else if content[index] == '\\' {
				escaped = true
			} else if content[index] == '\'' {
				state = stateCode
			}
			index++
		case stateDoubleQuote:
			if escaped {
				escaped = false
			} else if content[index] == '\\' {
				escaped = true
			} else if content[index] == '"' {
				state = stateCode
			}
			index++
		case stateLineComment:
			if content[index] == '\n' || content[index] == '\r' {
				state = stateCode
			}
			index++
		case stateBlockComment:
			if index+1 < len(content) && content[index] == '*' && content[index+1] == '/' {
				state = stateCode
				index += 2
				continue
			}
			index++
		default:
			if index+1 < len(content) {
				switch content[index : index+2] {
				case "//":
					state = stateLineComment
					index += 2
					continue
				case "/*":
					state = stateBlockComment
					index += 2
					continue
				}
			}
			switch content[index] {
			case '\'':
				state = stateSingleQuote
				index++
				continue
			case '"':
				state = stateDoubleQuote
				index++
				continue
			case '#':
				state = stateLineComment
				index++
				continue
			}
			if strings.HasPrefix(content[index:], symbol) && hasPHPNameBoundaries(content, index, index+len(symbol)) {
				offsets = append(offsets, index)
				index += len(symbol)
				continue
			}
			index++
		}
	}
	return offsets
}

func hasPHPNameBoundaries(content string, start int, end int) bool {
	before := start - 1
	if before >= 0 && content[before] == '\\' {
		before--
	}
	return isPHPNameBoundary(content, before) && isPHPNameBoundary(content, end)
}

func isPHPNameBoundary(content string, index int) bool {
	if index < 0 || index >= len(content) {
		return true
	}

	character := content[index]
	return !(character == '\\' || character == '_' || character >= '0' && character <= '9' || character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z')
}

func lineForOffset(content string, offset int) int {
	if offset < 0 || offset > len(content) {
		return 0
	}
	return strings.Count(content[:offset], "\n") + 1
}

func renderStalePHPSymbolHits(hits []stalePHPSymbolHit) string {
	lines := make([]string, 0, len(hits))
	for _, hit := range hits {
		lines = append(lines, fmt.Sprintf("%s:%d %s", hit.File, hit.Line, hit.Symbol))
	}
	return strings.Join(lines, "\n")
}

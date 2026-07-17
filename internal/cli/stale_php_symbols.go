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

type stalePHPNamespaceImport struct {
	offset int
	symbol string
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
	namespaceImportHits, err := scanStalePHPMovedNamespaceImports(projectRoot, scanConfig, mappings)
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
		return result, err
	}
	hits = append(hits, namespaceImportHits...)
	sortStalePHPSymbolHits(hits)
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

	sortStalePHPSymbolHits(hits)
	return hits, nil
}

func sortStalePHPSymbolHits(hits []stalePHPSymbolHit) {
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].File == hits[j].File {
			if hits[i].Line == hits[j].Line {
				return hits[i].Symbol < hits[j].Symbol
			}
			return hits[i].Line < hits[j].Line
		}
		return hits[i].File < hits[j].File
	})
}

func scanStalePHPMovedNamespaceImports(projectRoot string, scanConfig config.Config, mappings []adapterproto.SymbolMapping) ([]stalePHPSymbolHit, error) {
	type movedNamespaceFile struct {
		file         string
		oldNamespace string
	}

	index := map[movedNamespaceFile]struct{}{}
	for _, mapping := range mappings {
		if !isPHPSymbolKind(mapping.Kind) || mapping.NewPath == "" || mapping.OldNamespace == "" || mapping.OldNamespace == mapping.NewNamespace {
			continue
		}
		index[movedNamespaceFile{file: mapping.NewPath, oldNamespace: mapping.OldNamespace}] = struct{}{}
	}

	var hits []stalePHPSymbolHit
	for entry := range index {
		if filepath.Ext(entry.file) != ".php" || !scanConfig.Allows(entry.file) {
			continue
		}

		contentBytes, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(entry.file)))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		content := string(contentBytes)
		for _, namespaceImport := range codeNamespaceUseImports(content, entry.oldNamespace) {
			hits = append(hits, stalePHPSymbolHit{
				File:   entry.file,
				Line:   lineForOffset(content, namespaceImport.offset),
				Symbol: namespaceImport.symbol,
			})
		}
	}

	sortStalePHPSymbolHits(hits)
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

func codeNamespaceUseImports(content string, namespace string) []stalePHPNamespaceImport {
	const (
		stateCode = iota
		stateSingleQuote
		stateDoubleQuote
		stateLineComment
		stateBlockComment
	)

	state := stateCode
	escaped := false
	var imports []stalePHPNamespaceImport
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
			if symbolStart, symbolEnd, ok := staleNamespaceUseImportRange(content, index, namespace); ok {
				imports = append(imports, stalePHPNamespaceImport{
					offset: symbolStart,
					symbol: content[symbolStart:symbolEnd],
				})
				index = symbolEnd
				continue
			}
			index++
		}
	}
	return imports
}

func staleNamespaceUseImportRange(content string, index int, namespace string) (int, int, bool) {
	if !strings.HasPrefix(content[index:], "use") || !isPHPNameBoundary(content, index-1) {
		return 0, 0, false
	}

	cursor := index + len("use")
	if cursor >= len(content) || !isPHPWhitespace(content[cursor]) {
		return 0, 0, false
	}
	cursor = skipPHPWhitespace(content, cursor)

	for _, modifier := range []string{"function", "const"} {
		if strings.HasPrefix(content[cursor:], modifier) && isPHPNameBoundary(content, cursor-1) && isPHPNameBoundary(content, cursor+len(modifier)) {
			cursor = skipPHPWhitespace(content, cursor+len(modifier))
			break
		}
	}

	prefix := namespace + "\\"
	if !strings.HasPrefix(content[cursor:], prefix) {
		return 0, 0, false
	}

	end := cursor + len(prefix)
	for end < len(content) && isPHPNameCharacter(content[end]) {
		end++
	}
	if end == cursor+len(prefix) {
		return 0, 0, false
	}

	return cursor, end, true
}

func hasPHPNameBoundaries(content string, start int, end int) bool {
	before := start - 1
	if before >= 0 && content[before] == '\\' {
		before--
	}
	return isPHPNameBoundary(content, before) && isPHPNameBoundary(content, end)
}

func isPHPNameCharacter(character byte) bool {
	return character == '\\' || character == '_' || character >= '0' && character <= '9' || character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z'
}

func isPHPNameBoundary(content string, index int) bool {
	if index < 0 || index >= len(content) {
		return true
	}

	return !isPHPNameCharacter(content[index])
}

func skipPHPWhitespace(content string, index int) int {
	for index < len(content) && isPHPWhitespace(content[index]) {
		index++
	}
	return index
}

func isPHPWhitespace(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
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

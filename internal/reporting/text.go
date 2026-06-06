package reporting

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func RenderText(writer io.Writer, result Result) error {
	lines := []string{
		"Refactor plan",
		fmt.Sprintf("Mode: %s", modeLabel(result.DryRun)),
	}
	if result.ProjectRoot != "" {
		lines = append(lines, fmt.Sprintf("Project root: %s", result.ProjectRoot))
	}
	lines = append(lines, fmt.Sprintf("Semantic rewrites: %s", semanticRewriteLabel(result)))
	lines = append(lines, summaryLine(result))

	lines = append(lines, "")
	lines = append(lines, "Files:")
	lines = append(lines, formatFileSummaries(result)...)

	checks := filterChecks(result.Validation)
	if len(checks) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Checks:")
		lines = append(lines, formatValidation(checks)...)
	}

	if len(result.Errors) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Errors:")
		lines = append(lines, formatMessages(result.Errors)...)
	}

	_, err := fmt.Fprintln(writer, strings.Join(lines, "\n"))
	return err
}

func semanticRewriteLabel(result Result) string {
	if len(result.AutoDetectedAdapters) == 0 {
		return "none"
	}

	adapters := append([]string(nil), result.AutoDetectedAdapters...)
	sort.Strings(adapters)
	return strings.Join(adapters, ", ")
}

func modeLabel(dryRun bool) string {
	if dryRun {
		return "dry"
	}

	return "apply"
}

func summaryLine(result Result) string {
	editedFiles := make(map[string]struct{}, len(result.Replacements))
	for _, replacement := range result.Replacements {
		editedFiles[replacement.File] = struct{}{}
	}

	return fmt.Sprintf(
		"Summary: %d move(s), %d edited file(s), %d warning(s)",
		len(result.Moves),
		len(editedFiles),
		len(result.Warnings),
	)
}

func formatFileSummaries(result Result) []string {
	summaries := buildFileSummaries(result)
	if len(summaries) == 0 {
		return []string{"  (none)"}
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].sortKey == summaries[j].sortKey {
			return summaries[i].display < summaries[j].display
		}
		return summaries[i].sortKey < summaries[j].sortKey
	})

	lines := make([]string, 0, len(summaries)*3)
	for _, summary := range summaries {
		lines = append(lines, fmt.Sprintf("  %s", summary.display))
		if summary.moveMeta != "" {
			lines = append(lines, fmt.Sprintf("    move: %s", summary.moveMeta))
		}
		for _, detail := range summary.details() {
			lines = append(lines, fmt.Sprintf("    %s", detail))
		}
	}

	return lines
}

type fileSummary struct {
	sortKey     string
	display     string
	moveMeta    string
	symbols     []string
	paths       []string
	editActions map[string]map[string]int
	warnings    []Message
}

func buildFileSummaries(result Result) []*fileSummary {
	summaries := map[string]*fileSummary{}

	for _, move := range result.Moves {
		summary := ensureSummary(summaries, move.OldPath, fmt.Sprintf("%s -> %s", move.OldPath, move.NewPath))
		tracked := "untracked"
		if move.Tracked {
			tracked = "tracked"
		}
		summary.moveMeta = fmt.Sprintf("%s, %s", tracked, move.Mover)
	}

	for _, mapping := range result.SymbolMappings {
		summary := ensureMoveAwareSummary(summaries, mapping.OldPath, mapping.NewPath)
		summary.symbols = append(summary.symbols, fmt.Sprintf("%s: %s -> %s", symbolMappingLabel(mapping), mapping.OldSymbol, mapping.NewSymbol))
	}

	for _, mapping := range result.PathMappings {
		summary := ensureMoveAwareSummary(summaries, mapping.OldPath, mapping.NewPath)
		summary.paths = append(summary.paths, fmt.Sprintf("%s: %s -> %s", pathMappingLabel(mapping), mapping.OldReference, mapping.NewReference))
	}

	for _, replacement := range result.Replacements {
		summary := ensureMoveAwareSummary(summaries, replacement.File, replacement.File)
		adapter := replacement.Adapter
		if adapter == "" {
			adapter = "core"
		}
		if summary.editActions == nil {
			summary.editActions = map[string]map[string]int{}
		}
		if _, ok := summary.editActions[adapter]; !ok {
			summary.editActions[adapter] = map[string]int{}
		}
		summary.editActions[adapter][replacementActionLabel(replacement)]++
	}

	for _, warning := range result.Warnings {
		key := warning.File
		if key == "" {
			key = warning.Message
		}
		summary := ensureMoveAwareSummary(summaries, key, key)
		summary.warnings = append(summary.warnings, warning)
	}

	items := make([]*fileSummary, 0, len(summaries))
	for _, summary := range summaries {
		sort.Strings(summary.symbols)
		sort.Strings(summary.paths)
		sort.Slice(summary.warnings, func(i, j int) bool {
			if summary.warnings[i].File == summary.warnings[j].File {
				return summary.warnings[i].Line < summary.warnings[j].Line
			}
			return summary.warnings[i].File < summary.warnings[j].File
		})
		items = append(items, summary)
	}
	return items
}

func ensureSummary(summaries map[string]*fileSummary, key string, display string) *fileSummary {
	if summary, ok := summaries[key]; ok {
		if summary.display == key && display != key {
			summary.display = display
		}
		return summary
	}

	summary := &fileSummary{
		sortKey: key,
		display: display,
	}
	summaries[key] = summary
	return summary
}

func ensureMoveAwareSummary(summaries map[string]*fileSummary, oldKey string, displayKey string) *fileSummary {
	if summary, ok := summaries[oldKey]; ok {
		return summary
	}
	return ensureSummary(summaries, oldKey, displayKey)
}

func (s *fileSummary) details() []string {
	lines := make([]string, 0, len(s.symbols)+len(s.paths)+len(s.warnings)+4)

	for _, symbol := range s.symbols {
		lines = append(lines, symbol)
	}
	for _, path := range s.paths {
		lines = append(lines, path)
	}

	adapters := make([]string, 0, len(s.editActions))
	for adapter := range s.editActions {
		adapters = append(adapters, adapter)
	}
	sort.Strings(adapters)
	for _, adapter := range adapters {
		actions := make([]string, 0, len(s.editActions[adapter]))
		for action, count := range s.editActions[adapter] {
			if count == 1 {
				actions = append(actions, action)
				continue
			}
			actions = append(actions, fmt.Sprintf("%s x%d", action, count))
		}
		sort.Strings(actions)
		lines = append(lines, fmt.Sprintf("edits (%s): %s", adapter, strings.Join(actions, ", ")))
	}

	for _, warning := range s.warnings {
		if warning.Line > 0 {
			lines = append(lines, fmt.Sprintf("warning (line %d): %s", warning.Line, warning.Message))
			continue
		}
		lines = append(lines, fmt.Sprintf("warning: %s", warning.Message))
	}

	return lines
}

func symbolMappingLabel(mapping SymbolMapping) string {
	switch mapping.Kind {
	case "module":
		return "python module"
	case "package":
		return "go package"
	case "go-type":
		return "go type"
	case "go-function":
		return "go function"
	case "go-const":
		return "go const"
	case "go-var":
		return "go var"
	default:
		return "php symbol"
	}
}

func pathMappingLabel(mapping PathMapping) string {
	switch mapping.Kind {
	case "go-import-path":
		return "go import path"
	case "twig-template", "twig-template-directory":
		return "template reference"
	default:
		return "path reference"
	}
}

func formatMessages(messages []Message) []string {
	lines := make([]string, 0, len(messages))
	for _, message := range messages {
		location := message.File
		if location != "" && message.Line > 0 {
			location = fmt.Sprintf("%s:%d", location, message.Line)
		}
		if location != "" {
			lines = append(lines, fmt.Sprintf("  %s %s", location, message.Message))
			continue
		}
		lines = append(lines, fmt.Sprintf("  %s", message.Message))
	}
	return lines
}

func formatValidation(results []ValidationResult) []string {
	lines := make([]string, 0, len(results))
	for _, result := range results {
		if result.Name == "validation" {
			lines = append(lines, fmt.Sprintf("  %s", result.Message))
			continue
		}
		lines = append(lines, fmt.Sprintf("  %s: %s", result.Name, result.Message))
	}
	return lines
}

func filterChecks(results []ValidationResult) []ValidationResult {
	checks := make([]ValidationResult, 0, len(results))
	for _, result := range results {
		if result.Name == "replacement validation" {
			continue
		}
		checks = append(checks, result)
	}
	return checks
}

func replacementActionLabel(replacement ReplacementReport) string {
	switch replacement.Reason {
	case "php-namespace-declaration":
		return "namespace declaration"
	case "php-use-statement":
		return "use statement"
	case "php-fully-qualified-class-name":
		return "fully qualified class reference"
	case "php-class-constant":
		return "class constant reference"
	case "php-docblock-var":
		return "@var docblock"
	case "php-docblock-param":
		return "@param docblock"
	case "php-docblock-return":
		return "@return docblock"
	case "php-docblock-throws":
		return "@throws docblock"
	case "php-attribute-class-reference":
		return "attribute class reference"
	case "php-typed-property":
		return "typed property"
	case "php-method-parameter-type":
		return "parameter type"
	case "php-method-return-type":
		return "return type"
	case "go-import-path":
		return "import path"
	case "go-package-declaration":
		return "package declaration"
	case "go-package-qualifier":
		return "package qualifier"
	case "go-symbol-declaration":
		return "symbol declaration"
	case "go-local-symbol-reference":
		return "local symbol reference"
	case "go-imported-symbol-reference":
		return "imported symbol reference"
	}

	rule := replacement.Rule
	if rule != "" {
		if index := strings.LastIndex(rule, `\`); index >= 0 {
			rule = rule[index+1:]
		}
		rule = strings.TrimSuffix(rule, "ReplacementRule")
		rule = strings.TrimSuffix(rule, "Rule")
		if rule != "" {
			return splitCamel(rule)
		}
	}

	return strings.ReplaceAll(replacement.Reason, "-", " ")
}

func splitCamel(value string) string {
	if value == "" {
		return value
	}

	var builder strings.Builder
	for index, r := range value {
		if index > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteByte(' ')
		}
		builder.WriteRune(r)
	}

	return strings.ToLower(builder.String())
}

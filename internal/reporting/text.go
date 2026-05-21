package reporting

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Message struct {
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

type MoveReport struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
	Tracked bool   `json:"tracked"`
	Mover   string `json:"mover"`
}

type SymbolMapping struct {
	Kind      string `json:"kind"`
	OldPath   string `json:"oldPath"`
	NewPath   string `json:"newPath"`
	OldSymbol string `json:"oldSymbol"`
	NewSymbol string `json:"newSymbol"`
}

type PathMapping struct {
	Kind         string `json:"kind"`
	OldPath      string `json:"oldPath"`
	NewPath      string `json:"newPath"`
	OldReference string `json:"oldReference"`
	NewReference string `json:"newReference"`
}

type EditedFile struct {
	File         string `json:"file"`
	Replacements int    `json:"replacements"`
}

type ReplacementReport struct {
	File        string `json:"file"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
	Reason      string `json:"reason"`
	Rule        string `json:"rule,omitempty"`
	Adapter     string `json:"adapter,omitempty"`
	Replacement string `json:"replacement"`
}

type RuleResult struct {
	Rule         string `json:"rule"`
	Replacements int    `json:"replacements"`
}

type ValidationResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Stdout  string `json:"stdout,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
}

type Result struct {
	ProjectRoot            string              `json:"projectRoot,omitempty"`
	DryRun                 bool                `json:"dryRun"`
	Moves                  []MoveReport        `json:"moves"`
	AutoDetectedAdapters   []string            `json:"autoDetectedAdapters"`
	SymbolMappings         []SymbolMapping     `json:"symbolMappings"`
	PathMappings           []PathMapping       `json:"pathMappings"`
	EditedFiles            []EditedFile        `json:"editedFiles"`
	Replacements           []ReplacementReport `json:"replacements"`
	ReplacementRuleResults []RuleResult        `json:"replacementRuleResults"`
	Warnings               []Message           `json:"warnings"`
	Validation             []ValidationResult  `json:"validation"`
	Errors                 []Message           `json:"errors"`
	AdaptersDisabled       bool                `json:"adaptersDisabled,omitempty"`
}

func RenderText(writer io.Writer, result Result) error {
	lines := []string{
		"Refactor plan",
		fmt.Sprintf("Mode: %s", modeLabel(result.DryRun)),
	}
	if result.ProjectRoot != "" {
		lines = append(lines, fmt.Sprintf("Project root: %s", result.ProjectRoot))
	}
	lines = append(lines, fmt.Sprintf("Semantic rewrites: %s", semanticRewriteLabel(result)))

	lines = append(lines, "")
	lines = append(lines, "Moves:")
	lines = append(lines, formatMoves(result.Moves, result.SymbolMappings, result.PathMappings)...)

	lines = append(lines, "")
	lines = append(lines, "Edits:")
	lines = append(lines, formatEditSummaries(result.Replacements)...)

	if len(result.Warnings) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Warnings:")
		lines = append(lines, formatMessages(result.Warnings)...)
	}

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
	if result.AdaptersDisabled {
		return "disabled"
	}
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

func formatMoves(moves []MoveReport, symbolMappings []SymbolMapping, pathMappings []PathMapping) []string {
	if len(moves) == 0 {
		return []string{"  (none)"}
	}

	annotations := moveAnnotationsByPath(symbolMappings, pathMappings)
	lines := make([]string, 0, len(moves))
	for _, move := range moves {
		tracked := "untracked"
		if move.Tracked {
			tracked = "tracked"
		}
		lines = append(lines, fmt.Sprintf("  %s -> %s [%s, %s]", move.OldPath, move.NewPath, tracked, move.Mover))
		for _, annotation := range annotations[move.OldPath] {
			lines = append(lines, fmt.Sprintf("    %s", annotation))
		}
	}
	return lines
}

func moveAnnotationsByPath(symbolMappings []SymbolMapping, pathMappings []PathMapping) map[string][]string {
	annotations := map[string][]string{}
	for _, mapping := range symbolMappings {
		annotations[mapping.OldPath] = append(annotations[mapping.OldPath], fmt.Sprintf("php symbol: %s -> %s", mapping.OldSymbol, mapping.NewSymbol))
	}
	for _, mapping := range pathMappings {
		annotations[mapping.OldPath] = append(annotations[mapping.OldPath], fmt.Sprintf("twig reference: %s -> %s", mapping.OldReference, mapping.NewReference))
	}
	for path := range annotations {
		sort.Strings(annotations[path])
	}
	return annotations
}

func formatEditSummaries(replacements []ReplacementReport) []string {
	if len(replacements) == 0 {
		return []string{"  (none)"}
	}

	type fileSummary struct {
		byAdapter map[string]map[string]struct{}
	}

	summaries := map[string]*fileSummary{}
	for _, replacement := range replacements {
		summary, ok := summaries[replacement.File]
		if !ok {
			summary = &fileSummary{byAdapter: map[string]map[string]struct{}{}}
			summaries[replacement.File] = summary
		}
		adapter := replacement.Adapter
		if adapter == "" {
			adapter = "core"
		}
		if _, ok := summary.byAdapter[adapter]; !ok {
			summary.byAdapter[adapter] = map[string]struct{}{}
		}
		summary.byAdapter[adapter][replacementActionLabel(replacement)] = struct{}{}
	}

	files := make([]string, 0, len(summaries))
	for file := range summaries {
		files = append(files, file)
	}
	sort.Strings(files)

	lines := make([]string, 0, len(files)*2)
	for _, file := range files {
		lines = append(lines, fmt.Sprintf("  %s", file))
		adapters := make([]string, 0, len(summaries[file].byAdapter))
		for adapter := range summaries[file].byAdapter {
			adapters = append(adapters, adapter)
		}
		sort.Strings(adapters)
		for _, adapter := range adapters {
			actions := make([]string, 0, len(summaries[file].byAdapter[adapter]))
			for action := range summaries[file].byAdapter[adapter] {
				actions = append(actions, action)
			}
			sort.Strings(actions)
			lines = append(lines, fmt.Sprintf("    %s: %s", adapter, strings.Join(actions, ", ")))
		}
	}

	return lines
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
		lines = append(lines, fmt.Sprintf("  %s %s", result.Name, result.Message))
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

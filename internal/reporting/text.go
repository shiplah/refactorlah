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
	Worker      string `json:"worker,omitempty"`
	Adapter     string `json:"adapter,omitempty"`
	Replacement string `json:"replacement"`
}

type WorkerResult struct {
	Worker       string `json:"worker"`
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
	ProjectRoot              string              `json:"projectRoot,omitempty"`
	DryRun                   bool                `json:"dryRun"`
	Moves                    []MoveReport        `json:"moves"`
	AutoDetectedAdapters     []string            `json:"autoDetectedAdapters"`
	SymbolMappings           []SymbolMapping     `json:"symbolMappings"`
	PathMappings             []PathMapping       `json:"pathMappings"`
	EditedFiles              []EditedFile        `json:"editedFiles"`
	Replacements             []ReplacementReport `json:"replacements"`
	ReplacementWorkerResults []WorkerResult      `json:"replacementWorkerResults"`
	Warnings                 []Message           `json:"warnings"`
	Validation               []ValidationResult  `json:"validation"`
	Errors                   []Message           `json:"errors"`
	AdaptersDisabled         bool                `json:"adaptersDisabled,omitempty"`
}

func RenderText(writer io.Writer, result Result) error {
	lines := []string{
		"Refactor plan",
		fmt.Sprintf("Mode: %s", modeLabel(result.DryRun)),
	}
	if result.ProjectRoot != "" {
		lines = append(lines, fmt.Sprintf("Project root: %s", result.ProjectRoot))
	}

	lines = append(lines, "")
	lines = append(lines, "Moves:")
	lines = append(lines, formatMoves(result.Moves)...)
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Adapters: %s", formatAdapters(result.AutoDetectedAdapters, result.AdaptersDisabled)))

	if len(result.SymbolMappings) > 0 {
		lines = append(lines, "")
		lines = append(lines, "PHP symbols:")
		lines = append(lines, formatSymbolMappings(result.SymbolMappings)...)
	}

	if len(result.PathMappings) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Twig templates:")
		lines = append(lines, formatPathMappings(result.PathMappings)...)
	}

	lines = append(lines, "")
	lines = append(lines, "Edits:")
	lines = append(lines, formatEditedFiles(result.EditedFiles)...)

	if len(result.ReplacementWorkerResults) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Workers:")
		lines = append(lines, formatWorkerResults(result.ReplacementWorkerResults)...)
	}

	if len(result.Warnings) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Warnings:")
		lines = append(lines, formatMessages(result.Warnings)...)
	}

	if len(result.Validation) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Validation:")
		lines = append(lines, formatValidation(result.Validation)...)
	}

	if len(result.Errors) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Errors:")
		lines = append(lines, formatMessages(result.Errors)...)
	}

	_, err := fmt.Fprintln(writer, strings.Join(lines, "\n"))
	return err
}

func modeLabel(dryRun bool) string {
	if dryRun {
		return "dry"
	}

	return "apply"
}

func formatMoves(moves []MoveReport) []string {
	if len(moves) == 0 {
		return []string{"  (none)"}
	}

	lines := make([]string, 0, len(moves))
	for _, move := range moves {
		tracked := "untracked"
		if move.Tracked {
			tracked = "tracked"
		}
		lines = append(lines, fmt.Sprintf("  %s -> %s [%s, %s]", move.OldPath, move.NewPath, tracked, move.Mover))
	}
	return lines
}

func formatAdapters(adapters []string, disabled bool) string {
	if len(adapters) > 0 {
		return strings.Join(adapters, ", ")
	}
	if disabled {
		return "(disabled)"
	}
	return "(none)"
}

func formatSymbolMappings(mappings []SymbolMapping) []string {
	lines := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		lines = append(lines, fmt.Sprintf("  %s -> %s (%s)", mapping.OldSymbol, mapping.NewSymbol, mapping.OldPath))
	}
	return lines
}

func formatPathMappings(mappings []PathMapping) []string {
	lines := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		lines = append(lines, fmt.Sprintf("  %s -> %s (%s)", mapping.OldReference, mapping.NewReference, mapping.OldPath))
	}
	return lines
}

func formatEditedFiles(files []EditedFile) []string {
	if len(files) == 0 {
		return []string{"  (none)"}
	}

	lines := make([]string, 0, len(files))
	for _, file := range files {
		lines = append(lines, fmt.Sprintf("  %s (%d replacement(s))", file.File, file.Replacements))
	}
	return lines
}

func formatWorkerResults(results []WorkerResult) []string {
	lines := make([]string, 0, len(results))
	for _, result := range results {
		lines = append(lines, fmt.Sprintf("  %s: %d replacement(s)", result.Worker, result.Replacements))
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
		lines = append(lines, fmt.Sprintf("  %s: %s", result.Name, result.Message))
	}
	return lines
}

func sortMessages(messages []Message) {
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].File == messages[j].File {
			return messages[i].Message < messages[j].Message
		}
		return messages[i].File < messages[j].File
	})
}

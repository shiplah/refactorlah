package reporting

import (
	"fmt"
	"io"
	"sort"
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
	_, err := fmt.Fprintln(writer, "Move plan")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "========="); err != nil {
		return err
	}
	if result.ProjectRoot != "" {
		if _, err := fmt.Fprintf(writer, "\nProject root:\n  %s\n", result.ProjectRoot); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(writer, "\nFiles to move:"); err != nil {
		return err
	}
	if len(result.Moves) == 0 {
		if _, err := fmt.Fprintln(writer, "  (none)"); err != nil {
			return err
		}
	}
	for _, move := range result.Moves {
		tracked := "no"
		if move.Tracked {
			tracked = "yes"
		}
		if _, err := fmt.Fprintf(writer, "  %s\n    -> %s\n    tracked: %s\n    mover: %s\n", move.OldPath, move.NewPath, tracked, move.Mover); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(writer, "\nAuto-detected adapters:"); err != nil {
		return err
	}
	if len(result.AutoDetectedAdapters) == 0 {
		if result.AdaptersDisabled {
			if _, err := fmt.Fprintln(writer, "  (disabled)"); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintln(writer, "  (none)"); err != nil {
				return err
			}
		}
	}
	for _, adapter := range result.AutoDetectedAdapters {
		if _, err := fmt.Fprintf(writer, "  %s\n", adapter); err != nil {
			return err
		}
	}

	if len(result.SymbolMappings) > 0 {
		if _, err := fmt.Fprintln(writer, "\nPHP symbols:"); err != nil {
			return err
		}
		for _, mapping := range result.SymbolMappings {
			if _, err := fmt.Fprintf(writer, "  %s\n    %s\n    -> %s\n", mapping.OldPath, mapping.OldSymbol, mapping.NewSymbol); err != nil {
				return err
			}
		}
	}

	if len(result.PathMappings) > 0 {
		if _, err := fmt.Fprintln(writer, "\nTwig templates:"); err != nil {
			return err
		}
		for _, mapping := range result.PathMappings {
			if _, err := fmt.Fprintf(writer, "  %s\n    %s\n    -> %s\n", mapping.OldPath, mapping.OldReference, mapping.NewReference); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(writer, "\nFiles to edit:"); err != nil {
		return err
	}
	if len(result.EditedFiles) == 0 {
		if _, err := fmt.Fprintln(writer, "  (none)"); err != nil {
			return err
		}
	}
	for _, file := range result.EditedFiles {
		if _, err := fmt.Fprintf(writer, "  %s\n    %d replacement(s)\n", file.File, file.Replacements); err != nil {
			return err
		}
	}

	if len(result.ReplacementWorkerResults) > 0 {
		if _, err := fmt.Fprintln(writer, "\nReplacement workers:"); err != nil {
			return err
		}
		for _, worker := range result.ReplacementWorkerResults {
			if _, err := fmt.Fprintf(writer, "  %s: %d replacement(s)\n", worker.Worker, worker.Replacements); err != nil {
				return err
			}
		}
	}

	if len(result.Warnings) > 0 {
		if _, err := fmt.Fprintln(writer, "\nWarnings:"); err != nil {
			return err
		}
		for _, warning := range result.Warnings {
			if warning.File != "" {
				if _, err := fmt.Fprintf(writer, "  %s", warning.File); err != nil {
					return err
				}
				if warning.Line > 0 {
					if _, err := fmt.Fprintf(writer, ":%d", warning.Line); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintf(writer, "\n    %s\n", warning.Message); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(writer, "  %s\n", warning.Message); err != nil {
				return err
			}
		}
	}

	if len(result.Validation) > 0 {
		if _, err := fmt.Fprintln(writer, "\nValidation:"); err != nil {
			return err
		}
		for _, item := range result.Validation {
			if _, err := fmt.Fprintf(writer, "  %s: %s\n", item.Name, item.Message); err != nil {
				return err
			}
		}
	}

	if len(result.Errors) > 0 {
		if _, err := fmt.Fprintln(writer, "\nErrors:"); err != nil {
			return err
		}
		for _, failure := range result.Errors {
			if _, err := fmt.Fprintf(writer, "  %s\n", failure.Message); err != nil {
				return err
			}
		}
	}

	return nil
}

func sortMessages(messages []Message) {
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].File == messages[j].File {
			return messages[i].Message < messages[j].Message
		}
		return messages[i].File < messages[j].File
	})
}

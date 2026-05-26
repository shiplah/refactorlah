package reporting

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
}

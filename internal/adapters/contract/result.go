package contract

type SymbolMapping struct {
	Kind         string `json:"kind"`
	OldPath      string `json:"oldPath"`
	NewPath      string `json:"newPath"`
	OldSymbol    string `json:"oldSymbol"`
	NewSymbol    string `json:"newSymbol"`
	OldNamespace string `json:"oldNamespace,omitempty"`
	NewNamespace string `json:"newNamespace,omitempty"`
	ShortName    string `json:"shortName,omitempty"`
}

type PathMapping struct {
	Kind         string `json:"kind"`
	OldPath      string `json:"oldPath"`
	NewPath      string `json:"newPath"`
	OldReference string `json:"oldReference"`
	NewReference string `json:"newReference"`
}

type Replacement struct {
	File        string `json:"file"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
	Replacement string `json:"replacement"`
	Reason      string `json:"reason"`
	Rule        string `json:"rule,omitempty"`
	Adapter     string `json:"adapter,omitempty"`
}

type Warning struct {
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

type Check struct {
	Directory string   `json:"directory,omitempty"`
	Command   []string `json:"command"`
}

type AggregatedResponse struct {
	SymbolMappings []SymbolMapping
	PathMappings   []PathMapping
	Replacements   []Replacement
	Warnings       []Warning
	Checks         []Check
}

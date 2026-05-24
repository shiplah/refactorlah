package adapters

type Request struct {
	ProtocolVersion int            `json:"protocolVersion"`
	ProjectRoot     string         `json:"projectRoot"`
	OldPath         string         `json:"oldPath"`
	NewPath         string         `json:"newPath"`
	DryRun          bool           `json:"dryRun"`
	Moves           []Move         `json:"moves"`
	Options         RequestOptions `json:"options"`
}

type Move struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
	Tracked bool   `json:"tracked"`
}

type RequestOptions struct {
	IncludePHP  bool     `json:"includePhp"`
	IncludeTwig bool     `json:"includeTwig"`
	ScanInclude []string `json:"scanInclude,omitempty"`
	ScanExclude []string `json:"scanExclude,omitempty"`
}

type Response struct {
	ProtocolVersion int             `json:"protocolVersion"`
	Adapter         string          `json:"adapter"`
	SymbolMappings  []SymbolMapping `json:"symbolMappings"`
	PathMappings    []PathMapping   `json:"pathMappings"`
	Replacements    []Replacement   `json:"replacements"`
	Warnings        []Warning       `json:"warnings"`
	Errors          []string        `json:"errors"`
}

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

type AggregatedResponse struct {
	SymbolMappings []SymbolMapping
	PathMappings   []PathMapping
	Replacements   []Replacement
	Warnings       []Warning
}

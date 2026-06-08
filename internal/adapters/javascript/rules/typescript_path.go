package rules

import "github.com/NickSdot/refactorlah/internal/planning"

const (
	TypeScriptPathAliasReason    = "javascript-typescript-path-alias"
	TypeScriptPathAliasRuleName  = "javascript.TypeScriptPathAliasRule"
	TypeScriptPathTargetReason   = "javascript-typescript-path-target"
	TypeScriptPathTargetRuleName = "javascript.TypeScriptPathTargetRule"
)

type TypeScriptPathAliasRule struct{}

type TypeScriptPathTarget struct {
	Target string
}

type TypeScriptPathAmbiguity struct {
	Alias   string
	Targets []string
}

type TypeScriptPathTargetInput struct {
	ProjectRoot string
	File        string
	Content     []byte
	PathBase    string
	Targets     []TypeScriptPathTarget
	Moves       []planning.FileMove
}

type TypeScriptPathWarningInput struct {
	ProjectRoot string
	File        string
	PathBase    string
	Ambiguities []TypeScriptPathAmbiguity
	Moves       []planning.FileMove
}

type TypeScriptPathTargetRule struct{}
type TypeScriptPathWarningRule struct{}

package rules

import "refactorlah/internal/planning"

const (
	PackageImportsReason         = "javascript-package-imports"
	PackageImportsRuleName       = "javascript.PackageImportsRule"
	PackageImportTargetReason    = "javascript-package-import-target"
	PackageImportTargetRuleName  = "javascript.PackageImportTargetRule"
	PackageSelfReferenceReason   = "javascript-package-self-reference"
	PackageSelfReferenceRuleName = "javascript.PackageSelfReferenceRule"
)

type PackageImportTarget struct {
	Target string
}

type PackageConditionalImport struct {
	Key     string
	Targets []string
}

type PackageImportAliasRule struct{}
type PackageImportTargetRule struct{}
type PackageImportWarningRule struct{}

type PackageImportTargetInput struct {
	File    string
	Content []byte
	Targets []PackageImportTarget
	Moves   []planning.FileMove
}

type PackageImportWarningInput struct {
	File               string
	ConditionalImports []PackageConditionalImport
	Moves              []planning.FileMove
}

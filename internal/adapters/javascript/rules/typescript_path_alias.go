package rules

import (
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
)

func (r TypeScriptPathAliasRule) Collect(mappings []PathAliasMapping, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return PathAliasSpecifierRule{
		Reason: TypeScriptPathAliasReason,
		Rule:   TypeScriptPathAliasRuleName,
	}.Collect(mappings, moves)
}

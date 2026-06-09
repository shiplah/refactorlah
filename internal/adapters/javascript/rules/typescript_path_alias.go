package rules

import (
	"github.com/shiplah/refactorlah/internal/adapters/staticimports"
	"github.com/shiplah/refactorlah/internal/planning"
)

func (r TypeScriptPathAliasRule) Collect(mappings []PathAliasMapping, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return PathAliasSpecifierRule{
		Reason: TypeScriptPathAliasReason,
		Rule:   TypeScriptPathAliasRuleName,
	}.Collect(mappings, moves)
}

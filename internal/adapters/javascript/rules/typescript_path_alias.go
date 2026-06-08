package rules

import (
	"github.com/NickSdot/refactorlah/internal/adapters/staticimports"
	"github.com/NickSdot/refactorlah/internal/planning"
)

func (r TypeScriptPathAliasRule) Collect(mappings []PathAliasMapping, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return PathAliasSpecifierRule{
		Reason: TypeScriptPathAliasReason,
		Rule:   TypeScriptPathAliasRuleName,
	}.Collect(mappings, moves)
}

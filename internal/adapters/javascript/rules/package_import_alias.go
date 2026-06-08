package rules

import (
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
)

func (r PackageImportAliasRule) Collect(importMappings []PathAliasMapping, selfReferenceMappings []PathAliasMapping, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	rewrites := PathAliasSpecifierRule{
		Reason: PackageImportsReason,
		Rule:   PackageImportsRuleName,
	}.Collect(importMappings, moves)
	rewrites = append(rewrites, PathAliasSpecifierRule{
		Reason: PackageSelfReferenceReason,
		Rule:   PackageSelfReferenceRuleName,
	}.Collect(selfReferenceMappings, moves)...)
	return rewrites
}

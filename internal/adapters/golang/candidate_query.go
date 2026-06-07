package golang

import (
	"refactorlah/internal/adapters/scan"
)

func goCandidateQuery(packageMappings []packageMoveMapping, symbolMappings []symbolMoveMapping) scan.CandidateQuery {
	query := scan.CandidateQuery{
		Extensions: []string{".go"},
	}

	for _, mapping := range packageMappings {
		query.Needles = append(query.Needles, mapping.OldImport, mapping.OldPackage)
		for _, filePackage := range mapping.FilePackages {
			query.IncludePaths = append(query.IncludePaths, filePackage.OldPath)
		}
	}
	for _, mapping := range symbolMappings {
		query.Needles = append(query.Needles, mapping.OldImport, mapping.OldPackage, mapping.OldSymbol)
		query.IncludePaths = append(query.IncludePaths, mapping.OldPath)
	}

	return query
}

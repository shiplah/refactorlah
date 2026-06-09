//go:build cgo

package php

import (
	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/php/symfony/core"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/adapters/shared"
	"github.com/shiplah/refactorlah/internal/adapters/staticimports"
	"github.com/shiplah/refactorlah/internal/planning"
)

type ProjectPathCollector struct {
	staticImportScanner staticimports.Scanner
	assetMapperScanner  core.AssetMapperScanner
}

func NewProjectPathCollector() ProjectPathCollector {
	return ProjectPathCollector{
		staticImportScanner: staticimports.Scanner{},
		assetMapperScanner:  core.AssetMapperScanner{},
	}
}

func (c ProjectPathCollector) Collect(projectRoot string, composerRoot string, plan planning.MovePlan, containsStaticImport bool, scanIndex *scan.Index) ([]adapterproto.PathMapping, []adapterproto.Replacement, error) {
	var allReplacements []adapterproto.Replacement

	if containsStaticImport {
		staticFiles, err := scanIndex.CandidateFiles(composerRoot, scan.CandidateQuery{
			Extensions: []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".css"},
			Needles:    staticimports.CandidateNeedles(plan.Moves),
		})
		if err != nil {
			return nil, nil, err
		}
		staticReplacements, err := c.staticImportScanner.Scan(projectRoot, staticFiles, plan.Moves)
		if err != nil {
			return nil, nil, err
		}
		allReplacements = append(allReplacements, shared.ToAdapterReplacements(staticReplacements)...)
	}

	projectPathMappings := core.ProjectDirectoryPathMappings(plan)
	if len(projectPathMappings) == 0 {
		return nil, allReplacements, nil
	}

	yamlFiles, err := scanIndex.Files(composerRoot, ".yaml", ".yml")
	if err != nil {
		return nil, nil, err
	}
	assetMapperReplacements, err := c.assetMapperScanner.Scan(projectRoot, yamlFiles, projectPathMappings)
	if err != nil {
		return nil, nil, err
	}
	allReplacements = append(allReplacements, shared.ToAdapterReplacements(assetMapperReplacements)...)

	return projectPathMappings, allReplacements, nil
}

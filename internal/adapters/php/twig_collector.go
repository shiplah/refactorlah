//go:build cgo

package php

import (
	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/php/symfony/twig"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
	"github.com/NickSdot/refactorlah/internal/adapters/shared"
	"github.com/NickSdot/refactorlah/internal/planning"
)

type TwigCollector struct {
	configReader twig.ConfigReader
	mapper       twig.TemplateMapper
	ruleRegistry twig.RuleRegistry
}

func NewTwigCollector() TwigCollector {
	return TwigCollector{
		configReader: twig.ConfigReader{},
		mapper:       twig.TemplateMapper{},
		ruleRegistry: twig.NewRuleRegistry(),
	}
}

func (c TwigCollector) Collect(projectRoot string, composerRoot string, plan planning.MovePlan, scanIndex *scan.Index) ([]adapterproto.PathMapping, []adapterproto.Replacement, []adapterproto.Warning, error) {
	configuration, err := c.configReader.ReadFromConfigRoot(projectRoot, composerRoot)
	if err != nil {
		return nil, nil, nil, err
	}

	pathMappings := c.mapper.DeriveMappings(plan.Moves, configuration)
	if len(pathMappings) == 0 {
		return nil, nil, nil, nil
	}

	phpConfigFiles, err := scanIndex.Files(composerRoot, ".php", ".yaml", ".yml")
	if err != nil {
		return nil, nil, nil, err
	}
	twigFiles, err := scanIndex.Files(composerRoot, ".twig")
	if err != nil {
		return nil, nil, nil, err
	}

	twigReplacements, warnings, err := c.ruleRegistry.Scan(projectRoot, phpConfigFiles, twigFiles, pathMappings)
	if err != nil {
		return nil, nil, nil, err
	}

	return pathMappings, shared.ToAdapterReplacements(twigReplacements), warnings, nil
}

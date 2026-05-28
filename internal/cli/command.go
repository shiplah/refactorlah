package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"refactorlah/internal/adapters"
	"refactorlah/internal/config"
	"refactorlah/internal/git"
	"refactorlah/internal/planning"
	"refactorlah/internal/project"
	"refactorlah/internal/replacements"
	"refactorlah/internal/reporting"
	"refactorlah/internal/validation"
)

type Command struct {
	rootDetector     *project.RootDetector
	pathResolver     *project.PathResolver
	gitRepository    *git.Repository
	planner          *planning.Planner
	detector         *adapters.AutoDetector
	discovery        *adapters.Discovery
	invoker          *adapters.Invoker
	validator        *replacements.Validator
	applier          *replacements.Applier
	reportBuilder    *reporting.Builder
	validationRunner *validation.Runner
}

func NewCommand() *Command {
	gitRepo := git.NewRepository()
	return &Command{
		rootDetector:     project.NewRootDetector(gitRepo),
		pathResolver:     project.NewPathResolver(),
		gitRepository:    gitRepo,
		planner:          planning.NewPlanner(),
		detector:         adapters.NewAutoDetector(),
		discovery:        adapters.NewDiscovery(),
		invoker:          adapters.NewInvoker(),
		validator:        replacements.NewValidator(),
		applier:          replacements.NewApplier(),
		reportBuilder:    reporting.NewBuilder(),
		validationRunner: validation.NewRunner(),
	}
}

func (c *Command) Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	options, err := ParseOptions(args, stderr)
	if err != nil {
		if errors.Is(err, ErrHelpRequested) {
			WriteUsage(stdout)
			return ExitSuccess
		}

		var usageErr *UsageError
		if errors.As(err, &usageErr) {
			WriteUsageError(stderr, usageErr.Message)
			return ExitInvalidArguments
		}

		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidArguments
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: determine working directory: %v\n", err)
		return ExitGeneralFailure
	}

	result, exitCode := c.runWithOptions(ctx, cwd, options, stderr)
	if renderErr := c.renderResult(stdout, stderr, options, result); renderErr != nil {
		fmt.Fprintf(stderr, "error: render result: %v\n", renderErr)
		return ExitGeneralFailure
	}

	return exitCode
}

func (c *Command) runWithOptions(ctx context.Context, cwd string, options Options, stderr io.Writer) (reporting.Result, int) {
	rootInfo, err := c.rootDetector.Detect(ctx, cwd)
	if err != nil {
		return reporting.Result{
			DryRun: options.DryRun,
			Errors: []reporting.Message{{Message: err.Error()}},
		}, mapErrorToExitCode(err)
	}

	moveRequests, err := c.resolveMoveRequests(cwd, rootInfo.ProjectRoot, options)
	if err != nil {
		return reporting.Result{
			ProjectRoot: rootInfo.ProjectRoot,
			DryRun:      options.DryRun,
			Errors:      []reporting.Message{{Message: err.Error()}},
		}, ExitInvalidArguments
	}

	lockOptions := git.LockOptions{Writer: stderr}
	var applyLock *git.WorktreeLock
	if rootInfo.InGitRepo && options.Apply {
		applyLock, err = c.gitRepository.AcquireApplyLock(ctx, rootInfo.ProjectRoot, lockOptions)
		if err != nil {
			return reporting.Result{
				ProjectRoot: rootInfo.ProjectRoot,
				DryRun:      false,
				Errors:      []reporting.Message{{Message: err.Error()}},
			}, ExitMoveConflict
		}
		defer func() {
			_ = applyLock.Release()
		}()
	}

	if rootInfo.InGitRepo && options.Apply && options.RequireCleanWorktree {
		dirty, err := c.gitRepository.IsDirty(ctx, rootInfo.ProjectRoot)
		if err != nil {
			return reporting.Result{
				ProjectRoot: rootInfo.ProjectRoot,
				DryRun:      false,
				Errors:      []reporting.Message{{Message: err.Error()}},
			}, ExitGeneralFailure
		}
		if dirty {
			return reporting.Result{
				ProjectRoot: rootInfo.ProjectRoot,
				DryRun:      false,
				Errors: []reporting.Message{{
					Message: "git working tree is dirty; rerun without --require-clean-worktree to continue",
				}},
			}, ExitUnsafeWorkingTree
		}
	}

	plan, err := c.planner.BuildMany(ctx, rootInfo.ProjectRoot, moveRequests, trackingFunc(ctx, c.gitRepository, rootInfo))
	if err != nil {
		return reporting.Result{
			ProjectRoot: rootInfo.ProjectRoot,
			DryRun:      options.DryRun,
			Errors:      []reporting.Message{{Message: err.Error()}},
		}, mapErrorToExitCode(err)
	}

	validationRoot := rootInfo.ProjectRoot
	if composerRoot, found, err := project.FindComposerRootForPaths(rootInfo.ProjectRoot, moveRequestPaths(moveRequests)); err == nil && found {
		validationRoot = composerRoot
	} else if pythonRoot, found, err := project.FindPythonRootForPaths(rootInfo.ProjectRoot, moveRequestPaths(moveRequests)); err == nil && found {
		validationRoot = pythonRoot
	}

	scanConfig, err := config.NewLoader().Load(rootInfo.ProjectRoot, cwd)
	if err != nil {
		return reporting.Result{
			ProjectRoot: rootInfo.ProjectRoot,
			DryRun:      options.DryRun,
			Moves:       c.reportBuilder.MoveReports(plan),
			Errors:      []reporting.Message{{Message: err.Error()}},
		}, ExitGeneralFailure
	}

	adapterSelection, discoveryWarnings, err := c.prepareAdapters(ctx, rootInfo.ProjectRoot, plan, options, scanConfig)
	if err != nil {
		return reporting.Result{
			ProjectRoot: rootInfo.ProjectRoot,
			DryRun:      options.DryRun,
			Moves:       c.reportBuilder.MoveReports(plan),
			Warnings:    discoveryWarnings,
			Errors:      []reporting.Message{{Message: err.Error()}},
		}, mapErrorToExitCode(err)
	}

	adapterOutput, adapterWarnings, err := c.runAdapters(ctx, rootInfo.ProjectRoot, plan, options, adapterSelection)
	if err != nil {
		return reporting.Result{
			ProjectRoot:          rootInfo.ProjectRoot,
			DryRun:               options.DryRun,
			Moves:                c.reportBuilder.MoveReports(plan),
			AutoDetectedAdapters: adapterSelection.Names(),
			Warnings:             append(discoveryWarnings, adapterWarnings...),
			Errors:               []reporting.Message{{Message: err.Error()}},
		}, mapErrorToExitCode(err)
	}
	adapterOutput.Replacements = replacements.Deduplicate(adapterOutput.Replacements)

	validationIssues, err := c.validator.Validate(rootInfo.ProjectRoot, adapterOutput.Replacements)
	if err != nil {
		return reporting.Result{
			ProjectRoot:          rootInfo.ProjectRoot,
			DryRun:               options.DryRun,
			Moves:                c.reportBuilder.MoveReports(plan),
			AutoDetectedAdapters: adapterSelection.Names(),
			SymbolMappings:       c.reportBuilder.SymbolMappings(adapterOutput.SymbolMappings),
			PathMappings:         c.reportBuilder.PathMappings(adapterOutput.PathMappings),
			Warnings:             append(append(discoveryWarnings, adapterWarnings...), warningMessages(adapterOutput.Warnings)...),
			Errors:               []reporting.Message{{Message: err.Error()}},
		}, mapErrorToExitCode(err)
	}

	report := reporting.Result{
		ProjectRoot:            rootInfo.ProjectRoot,
		DryRun:                 options.DryRun,
		Moves:                  c.reportBuilder.MoveReports(plan),
		AutoDetectedAdapters:   adapterSelection.Names(),
		SymbolMappings:         c.reportBuilder.SymbolMappings(adapterOutput.SymbolMappings),
		PathMappings:           c.reportBuilder.PathMappings(adapterOutput.PathMappings),
		EditedFiles:            c.reportBuilder.EditedFiles(adapterOutput.Replacements),
		Replacements:           c.reportBuilder.Replacements(adapterOutput.Replacements),
		ReplacementRuleResults: c.reportBuilder.RuleResults(adapterOutput.Replacements),
		Warnings:               append(append(discoveryWarnings, adapterWarnings...), warningMessages(adapterOutput.Warnings)...),
		Validation:             validationIssues,
	}

	if options.DryRun {
		return report, ExitSuccess
	}

	validationBaseline := c.validationRunner.Baseline(ctx, validationRoot, validation.RunOptions{
		SkipValidation: options.NoValidation,
		RunTests:       options.RunTests,
	})

	if err := c.gitRepository.MoveFiles(ctx, rootInfo.ProjectRoot, plan.Moves, lockOptions); err != nil {
		report.Errors = []reporting.Message{{Message: err.Error()}}
		return report, ExitMoveConflict
	}

	if err := c.applier.Apply(rootInfo.ProjectRoot, plan.TargetPaths(), adapterOutput.Replacements); err != nil {
		report.Errors = []reporting.Message{{Message: err.Error()}}
		return report, ExitReplacementConflict
	}

	validationResults, err := c.validationRunner.RunCompared(ctx, validationRoot, validation.RunOptions{
		SkipValidation: options.NoValidation,
		RunTests:       options.RunTests,
	}, validationBaseline)
	report.Validation = validationResults
	if err != nil {
		report.Errors = []reporting.Message{{Message: err.Error()}}
		return report, ExitValidationFailed
	}

	return report, ExitSuccess
}

func (c *Command) resolveMoveRequests(cwd string, projectRoot string, options Options) ([]planning.RequestedMove, error) {
	requests := options.MoveRequests
	if options.UseFile != "" {
		expanded, err := ReadMoveFile(cwd, options.UseFile)
		if err != nil {
			return nil, err
		}
		requests = expanded
	} else if len(requests) == 0 && options.OldPath != "" && options.NewPath != "" {
		requests = []planning.RequestedMove{{
			OldPath: options.OldPath,
			NewPath: options.NewPath,
		}}
	}

	resolved := make([]planning.RequestedMove, 0, len(requests))
	for _, request := range requests {
		oldPath, err := c.pathResolver.Resolve(projectRoot, request.OldPath)
		if err != nil {
			return nil, err
		}
		newPath, err := c.pathResolver.Resolve(projectRoot, request.NewPath)
		if err != nil {
			return nil, err
		}
		if oldPath == newPath {
			return nil, errors.New("old-path and new-path resolve to the same path")
		}
		resolved = append(resolved, planning.RequestedMove{
			OldPath: oldPath,
			NewPath: newPath,
		})
	}

	return expandWildcardRequests(projectRoot, resolved)
}

func moveRequestPaths(requests []planning.RequestedMove) []string {
	paths := make([]string, 0, len(requests)*2)
	for _, request := range requests {
		paths = append(paths, request.OldPath, request.NewPath)
	}

	return paths
}

func (c *Command) prepareAdapters(ctx context.Context, projectRoot string, plan planning.MovePlan, options Options, scanConfig config.Config) (adapters.Selection, []reporting.Message, error) {
	signals, err := c.detector.Detect(ctx, projectRoot, plan)
	if err != nil {
		return adapters.Selection{}, nil, err
	}

	selection := adapters.Selection{}
	warnings := []reporting.Message{}

	if signals.PHPRelevant {
		path, err := c.discovery.RequirePHPAdapter(ctx, projectRoot)
		if err != nil {
			return selection, warnings, err
		}

		selection.Adapters = append(selection.Adapters, adapters.Config{
			Name: "php",
			Path: path,
			Options: adapters.RequestOptions{
				IncludePHP:  signals.IncludePHP,
				IncludeTwig: signals.IncludeTwig,
				ScanInclude: scanConfig.Include,
				ScanExclude: scanConfig.Exclude,
			},
		})
	}

	if signals.PythonRelevant {
		path, err := c.discovery.RequirePythonAdapter(ctx, projectRoot)
		if err != nil {
			return selection, warnings, err
		}

		selection.Adapters = append(selection.Adapters, adapters.Config{
			Name: "python",
			Path: path,
			Options: adapters.RequestOptions{
				IncludePython: signals.IncludePython,
				ScanInclude:   scanConfig.Include,
				ScanExclude:   scanConfig.Exclude,
			},
		})
	}

	return selection, warnings, nil
}

func (c *Command) runAdapters(ctx context.Context, projectRoot string, plan planning.MovePlan, options Options, selection adapters.Selection) (adapters.AggregatedResponse, []reporting.Message, error) {
	if len(selection.Adapters) == 0 {
		return adapters.AggregatedResponse{}, nil, nil
	}

	response, err := c.invoker.Invoke(ctx, projectRoot, plan, options.DryRun, selection)
	if err != nil {
		return adapters.AggregatedResponse{}, nil, err
	}

	return response, nil, nil
}

func (c *Command) renderResult(stdout io.Writer, stderr io.Writer, options Options, result reporting.Result) error {
	switch options.Format {
	case FormatJSON:
		return reporting.RenderJSON(stdout, result)
	default:
		return reporting.RenderText(stdout, result)
	}
}

func trackingFunc(ctx context.Context, repo *git.Repository, rootInfo project.RootInfo) planning.TrackFunc {
	if !rootInfo.InGitRepo {
		return func(string) (bool, error) {
			return false, nil
		}
	}

	return func(path string) (bool, error) {
		return repo.IsTracked(ctx, rootInfo.ProjectRoot, path)
	}
}

func mapErrorToExitCode(err error) int {
	switch {
	case errors.Is(err, planning.ErrTargetExists):
		return ExitMoveConflict
	case errors.Is(err, replacements.ErrConflict):
		return ExitReplacementConflict
	case errors.Is(err, adapters.ErrAdapterFailure):
		return ExitAdapterFailure
	case errors.Is(err, validation.ErrValidationFailed):
		return ExitValidationFailed
	default:
		return ExitGeneralFailure
	}
}

func warningMessages(warnings []adapters.Warning) []reporting.Message {
	result := make([]reporting.Message, 0, len(warnings))
	for _, warning := range warnings {
		result = append(result, reporting.Message{
			File:    warning.File,
			Line:    warning.Line,
			Message: warning.Message,
		})
	}
	return result
}

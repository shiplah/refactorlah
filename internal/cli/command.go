package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/registry"
	"github.com/NickSdot/refactorlah/internal/config"
	"github.com/NickSdot/refactorlah/internal/git"
	"github.com/NickSdot/refactorlah/internal/planning"
	"github.com/NickSdot/refactorlah/internal/project"
	"github.com/NickSdot/refactorlah/internal/replacements"
	"github.com/NickSdot/refactorlah/internal/reporting"
	"github.com/NickSdot/refactorlah/internal/validation"
)

type Command struct {
	rootDetector     *project.RootDetector
	pathResolver     *project.PathResolver
	gitRepository    *git.Repository
	planner          *planning.Planner
	nativeAnalyzers  *registry.Registry
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
		nativeAnalyzers:  registry.NewRegistry(),
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
			WriteMoveUsage(stdout)
			return ExitSuccess
		}

		var usageErr *UsageError
		if errors.As(err, &usageErr) {
			WriteMoveUsageError(stderr, usageErr.Message)
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

	scanConfig, err := config.NewLoader().Load(rootInfo.ProjectRoot, cwd)
	if err != nil {
		return reporting.Result{
			ProjectRoot: rootInfo.ProjectRoot,
			DryRun:      options.DryRun,
			Moves:       c.reportBuilder.MoveReports(plan),
			Errors:      []reporting.Message{{Message: err.Error()}},
		}, ExitGeneralFailure
	}

	if err := validateMovePlanAllowedByConfig(plan, scanConfig); err != nil {
		return reporting.Result{
			ProjectRoot: rootInfo.ProjectRoot,
			DryRun:      options.DryRun,
			Moves:       c.reportBuilder.MoveReports(plan),
			Errors:      []reporting.Message{{Message: err.Error()}},
		}, ExitInvalidArguments
	}

	adapterOutput, semanticSources, err := c.runNativeAnalyzers(rootInfo.ProjectRoot, plan, scanConfig)
	if err != nil {
		return reporting.Result{
			ProjectRoot:          rootInfo.ProjectRoot,
			DryRun:               options.DryRun,
			Moves:                c.reportBuilder.MoveReports(plan),
			AutoDetectedAdapters: semanticSources,
			Errors:               []reporting.Message{{Message: err.Error()}},
		}, ExitAdapterFailure
	}

	adapterOutput.Replacements = replacements.Deduplicate(adapterOutput.Replacements)
	checks := validationChecks(adapterOutput.Checks, scanConfig.Checks)
	testChecks := validation.ChecksFromCommands(scanConfig.Tests)

	validationIssues, err := c.validator.Validate(rootInfo.ProjectRoot, adapterOutput.Replacements)
	if err != nil {
		return reporting.Result{
			ProjectRoot:          rootInfo.ProjectRoot,
			DryRun:               options.DryRun,
			Moves:                c.reportBuilder.MoveReports(plan),
			AutoDetectedAdapters: semanticSources,
			SymbolMappings:       c.reportBuilder.SymbolMappings(adapterOutput.SymbolMappings),
			PathMappings:         c.reportBuilder.PathMappings(adapterOutput.PathMappings),
			Warnings:             warningMessages(adapterOutput.Warnings),
			Errors:               []reporting.Message{{Message: err.Error()}},
		}, mapErrorToExitCode(err)
	}

	report := reporting.Result{
		ProjectRoot:            rootInfo.ProjectRoot,
		DryRun:                 options.DryRun,
		Moves:                  c.reportBuilder.MoveReports(plan),
		AutoDetectedAdapters:   semanticSources,
		SymbolMappings:         c.reportBuilder.SymbolMappings(adapterOutput.SymbolMappings),
		PathMappings:           c.reportBuilder.PathMappings(adapterOutput.PathMappings),
		EditedFiles:            c.reportBuilder.EditedFiles(adapterOutput.Replacements),
		Replacements:           c.reportBuilder.Replacements(adapterOutput.Replacements),
		ReplacementRuleResults: c.reportBuilder.RuleResults(adapterOutput.Replacements),
		Warnings:               warningMessages(adapterOutput.Warnings),
		Validation:             validationIssues,
	}

	if options.DryRun {
		report.Validation = append(report.Validation, c.validationRunner.Plan(checks, testChecks, validation.RunOptions{
			SkipValidation: options.NoValidation,
			RunTests:       options.RunTests,
		})...)
		return report, ExitSuccess
	}

	if err := c.gitRepository.MoveFiles(ctx, rootInfo.ProjectRoot, plan.Moves, lockOptions); err != nil {
		report.Errors = []reporting.Message{{Message: err.Error()}}
		return report, ExitMoveConflict
	}

	if err := c.applier.Apply(rootInfo.ProjectRoot, plan.TargetPaths(), adapterOutput.Replacements); err != nil {
		report.Errors = []reporting.Message{{Message: err.Error()}}
		return report, ExitReplacementConflict
	}

	validationResults, err := c.validationRunner.Run(ctx, rootInfo.ProjectRoot, checks, testChecks, validation.RunOptions{
		SkipValidation: options.NoValidation,
		RunTests:       options.RunTests,
	})
	report.Validation = append(validationIssues, validationResults...)
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
		oldPath, newPath, err := c.pathResolver.ResolveMove(projectRoot, cwd, request.OldPath, request.NewPath)
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

func (c *Command) runNativeAnalyzers(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (contract.AggregatedResponse, []string, error) {
	return c.nativeAnalyzers.Analyze(projectRoot, plan, scanConfig)
}

func validateMovePlanAllowedByConfig(plan planning.MovePlan, scanConfig config.Config) error {
	for _, move := range plan.Moves {
		if !scanConfig.Allows(move.OldPath) {
			return fmt.Errorf("move source %q is excluded by .refactorlah.json", move.OldPath)
		}
		if !scanConfig.Allows(move.NewPath) {
			return fmt.Errorf("move target %q is excluded by .refactorlah.json", move.NewPath)
		}
	}
	return nil
}

func validationChecks(adapterChecks []contract.Check, configuredCommands [][]string) []validation.Check {
	checks := make([]validation.Check, 0, len(adapterChecks)+len(configuredCommands))
	for _, check := range adapterChecks {
		if len(check.Command) == 0 {
			continue
		}
		checks = append(checks, validation.Check{
			Directory: check.Directory,
			Command:   append([]string(nil), check.Command...),
		})
	}
	checks = append(checks, validation.ChecksFromCommands(configuredCommands)...)
	return checks
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
	case errors.Is(err, validation.ErrValidationFailed):
		return ExitValidationFailed
	default:
		return ExitGeneralFailure
	}
}

func warningMessages(warnings []contract.Warning) []reporting.Message {
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

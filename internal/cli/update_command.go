package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"refactorlah/internal/buildinfo"
	"refactorlah/internal/selfupdate"
)

type UpdateCommand struct {
	stdin      io.Reader
	newUpdater func() (*selfupdate.Updater, error)
}

func NewUpdateCommand() *UpdateCommand {
	return &UpdateCommand{
		stdin:      os.Stdin,
		newUpdater: selfupdate.NewUpdater,
	}
}

func (c *UpdateCommand) Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var checkOnly bool
	var yes bool
	var targetVersion string
	var jsonOutput bool

	flags.BoolVar(&checkOnly, "check", false, "check whether a newer version is available")
	flags.BoolVar(&yes, "yes", false, "apply the update without prompting")
	flags.StringVar(&targetVersion, "to", "", "update to an explicit release tag")
	flags.BoolVar(&jsonOutput, "json", false, "print JSON output")

	if err := flags.Parse(args); err != nil {
		WriteCommandUsageError(stderr, err.Error())
		WriteUpdateUsage(stderr)
		return ExitInvalidArguments
	}
	if flags.NArg() != 0 {
		WriteCommandUsageError(stderr, "update does not accept positional arguments")
		WriteUpdateUsage(stderr)
		return ExitInvalidArguments
	}
	if jsonOutput && !checkOnly && !yes {
		WriteCommandUsageError(stderr, "--json requires --check or --yes")
		WriteUpdateUsage(stderr)
		return ExitInvalidArguments
	}

	updaterFactory := c.newUpdater
	if updaterFactory == nil {
		updaterFactory = selfupdate.NewUpdater
	}

	updater, err := updaterFactory()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitGeneralFailure
	}
	updater.Stdout = stdout
	updater.Stderr = stderr

	options := selfupdate.CheckOptions{TargetVersion: targetVersion}
	plan, err := updater.Plan(ctx, options)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitGeneralFailure
	}

	if checkOnly {
		return renderUpdateCheckResult(stdout, stderr, plan.CheckResult, jsonOutput)
	}
	if !plan.CheckResult.UpdateAvailable {
		return renderUpdateApplyResult(stdout, stderr, selfupdate.ApplyResult{CheckResult: plan.CheckResult}, jsonOutput)
	}

	if !yes {
		if !confirmUpdate(c.stdin, stdout, plan.CheckResult) {
			_, _ = fmt.Fprintln(stdout, "Update cancelled.")
			return ExitSuccess
		}
	}

	applyResult, err := updater.ApplyPlan(ctx, plan)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitGeneralFailure
	}

	return renderUpdateApplyResult(stdout, stderr, applyResult, jsonOutput)
}

func confirmUpdate(stdin io.Reader, stdout io.Writer, result selfupdate.CheckResult) bool {
	_, _ = fmt.Fprintf(stdout, "%s [y/N]: ", updatePrompt(result))

	reader := bufio.NewReader(stdin)
	response, err := reader.ReadString('\n')
	if err != nil && len(response) == 0 {
		return false
	}

	answer := strings.ToLower(strings.TrimSpace(response))
	return answer == "y" || answer == "yes"
}

func updatePrompt(result selfupdate.CheckResult) string {
	if result.Downgrade {
		return fmt.Sprintf("Install %s over current %s at %s?", result.TargetVersion, result.CurrentVersion, result.ExecutablePath)
	}

	if promptsBeforeReplacingSourceBuild(result.CurrentDistribution) {
		return fmt.Sprintf(
			"Replace the current %s build (%s) at %s with published release %s?",
			result.CurrentDistribution,
			result.CurrentVersion,
			result.ExecutablePath,
			result.TargetVersion,
		)
	}

	return fmt.Sprintf("Install %s at %s?", result.TargetVersion, result.ExecutablePath)
}

func promptsBeforeReplacingSourceBuild(distribution string) bool {
	return distribution == buildinfo.DistributionSourceInstall || distribution == buildinfo.DistributionDev
}

func renderUpdateCheckResult(stdout io.Writer, stderr io.Writer, result selfupdate.CheckResult, jsonOutput bool) int {
	if jsonOutput {
		if err := writeJSONOutput(stdout, result); err != nil {
			fmt.Fprintf(stderr, "error: write update output: %v\n", err)
			return ExitGeneralFailure
		}
		return ExitSuccess
	}

	switch {
	case result.UpdateAvailable:
		_, _ = fmt.Fprintf(stdout, "Update available: %s -> %s\n", result.CurrentVersion, result.TargetVersion)
		_, _ = fmt.Fprintf(stdout, "Asset: %s\n", result.AssetName)
		_, _ = fmt.Fprintf(stdout, "Release: %s\n", result.ReleaseURL)
	case result.Downgrade:
		_, _ = fmt.Fprintf(stdout, "Current version %s is newer than published release %s\n", result.CurrentVersion, result.TargetVersion)
	default:
		_, _ = fmt.Fprintf(stdout, "refactorlah is up to date (%s)\n", result.CurrentVersion)
	}

	return ExitSuccess
}

func renderUpdateApplyResult(stdout io.Writer, stderr io.Writer, result selfupdate.ApplyResult, jsonOutput bool) int {
	if jsonOutput {
		if err := writeJSONOutput(stdout, result); err != nil {
			fmt.Fprintf(stderr, "error: write update output: %v\n", err)
			return ExitGeneralFailure
		}
		return ExitSuccess
	}

	switch {
	case result.Staged:
		_, _ = fmt.Fprintf(stdout, "Update to %s staged.\n", result.TargetVersion)
		_, _ = fmt.Fprintln(stdout, "Restart refactorlah to use the new version.")
	case result.Updated:
		_, _ = fmt.Fprintf(stdout, "Updated refactorlah to %s.\n", result.TargetVersion)
	case result.Downgrade:
		_, _ = fmt.Fprintf(stdout, "Current version %s is newer than published release %s\n", result.CurrentVersion, result.TargetVersion)
	case result.UpToDate:
		_, _ = fmt.Fprintf(stdout, "refactorlah is up to date (%s)\n", result.CurrentVersion)
	default:
		_, _ = fmt.Fprintf(stdout, "refactorlah is up to date (%s)\n", result.CurrentVersion)
	}

	return ExitSuccess
}

func WriteUpdateUsage(writer io.Writer) {
	_, _ = io.WriteString(writer, "Usage:\n")
	_, _ = io.WriteString(writer, "  refactorlah update [--check] [--yes] [--to TAG] [--json]\n")
}

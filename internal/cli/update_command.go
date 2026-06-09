package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shiplah/refactorlah/internal/selfupdate"
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
		if errors.Is(err, flag.ErrHelp) {
			WriteUpdateUsage(stdout)
			return ExitSuccess
		}
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
	if !plan.CheckResult.SelfUpdateSupported {
		return renderUnsupportedSelfUpdate(stdout, stderr, plan.CheckResult, jsonOutput)
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

	return fmt.Sprintf("Install %s at %s?", result.TargetVersion, result.ExecutablePath)
}

func renderUpdateCheckResult(stdout io.Writer, stderr io.Writer, result selfupdate.CheckResult, jsonOutput bool) int {
	if jsonOutput {
		if err := writeJSONOutput(stdout, result); err != nil {
			fmt.Fprintf(stderr, "error: write update output: %v\n", err)
			return ExitGeneralFailure
		}
		return ExitSuccess
	}

	if !result.SelfUpdateSupported {
		renderUnsupportedSelfUpdateCheck(stdout, result)
		return ExitSuccess
	}

	switch {
	case result.Downgrade:
		_, _ = fmt.Fprintf(stdout, "Current version %s is newer than published release %s\n", result.CurrentVersion, result.TargetVersion)
	case result.UpdateAvailable:
		_, _ = fmt.Fprintf(stdout, "Update available: %s -> %s\n", result.CurrentVersion, result.TargetVersion)
		_, _ = fmt.Fprintf(stdout, "Asset: %s\n", result.AssetName)
		_, _ = fmt.Fprintf(stdout, "Release: %s\n", result.ReleaseURL)
	default:
		_, _ = fmt.Fprintf(stdout, "refactorlah is up to date (%s)\n", result.CurrentVersion)
	}

	return ExitSuccess
}

func renderUnsupportedSelfUpdateCheck(stdout io.Writer, result selfupdate.CheckResult) {
	_, _ = fmt.Fprintf(stdout, "Current build: %s (%s)\n", result.CurrentVersion, result.CurrentDistribution)
	if result.TargetVersion != "" {
		if result.UpdateAvailable {
			_, _ = fmt.Fprintf(stdout, "Published release available: %s\n", result.TargetVersion)
		} else {
			_, _ = fmt.Fprintf(stdout, "Published release: %s\n", result.TargetVersion)
		}
	}
	_, _ = fmt.Fprintln(stdout, "Self-update is only available for GitHub release binaries.")
	if result.UpdateInstructions != "" {
		_, _ = fmt.Fprintln(stdout, result.UpdateInstructions)
	}
}

func renderUnsupportedSelfUpdate(stdout io.Writer, stderr io.Writer, result selfupdate.CheckResult, jsonOutput bool) int {
	if jsonOutput {
		applyResult := selfupdate.ApplyResult{CheckResult: result}
		if err := writeJSONOutput(stdout, applyResult); err != nil {
			fmt.Fprintf(stderr, "error: write update output: %v\n", err)
			return ExitGeneralFailure
		}
		return ExitGeneralFailure
	}

	renderUnsupportedSelfUpdateCheck(stdout, result)
	return ExitGeneralFailure
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
	_, _ = io.WriteString(writer, "\nOptions:\n")
	WriteUpdateOptions(writer, "  ")
}

func WriteUpdateOptions(writer io.Writer, indent string) {
	_, _ = fmt.Fprintf(writer, "%s--check  Check whether a newer version is available\n", indent)
	_, _ = fmt.Fprintf(writer, "%s--yes    Apply the update without prompting\n", indent)
	_, _ = fmt.Fprintf(writer, "%s--to TAG  Update to an explicit release tag\n", indent)
	_, _ = fmt.Fprintf(writer, "%s--json   Print JSON output\n", indent)
	_, _ = fmt.Fprintf(writer, "%s--help   Show this help\n", indent)
}

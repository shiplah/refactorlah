package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

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
		writeCommandUsageError(stderr, err.Error())
		WriteUpdateUsage(stderr)
		return ExitInvalidArguments
	}
	if flags.NArg() != 0 {
		writeCommandUsageError(stderr, "update does not accept positional arguments")
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
	if checkOnly {
		result, err := updater.Check(ctx, options)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return ExitGeneralFailure
		}
		return renderUpdateCheckResult(stdout, stderr, result, jsonOutput)
	}

	result, err := updater.Check(ctx, options)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitGeneralFailure
	}
	if !result.UpdateAvailable {
		return renderUpdateApplyResult(stdout, stderr, selfupdate.ApplyResult{CheckResult: result}, jsonOutput)
	}

	if !yes {
		if !confirmUpdate(c.stdin, stdout, result) {
			_, _ = fmt.Fprintln(stdout, "Update cancelled.")
			return ExitSuccess
		}
	}

	applyResult, err := updater.Apply(ctx, options)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitGeneralFailure
	}

	return renderUpdateApplyResult(stdout, stderr, applyResult, jsonOutput)
}

func confirmUpdate(stdin io.Reader, stdout io.Writer, result selfupdate.CheckResult) bool {
	if result.Downgrade {
		_, _ = fmt.Fprintf(stdout, "Install %s over current %s at %s? [y/N]: ", result.TargetVersion, result.CurrentVersion, result.ExecutablePath)
	} else if result.CurrentDistribution == "source-install" || result.CurrentDistribution == "dev" {
		_, _ = fmt.Fprintf(stdout, "Replace the current %s build (%s) at %s with published release %s? [y/N]: ", result.CurrentDistribution, result.CurrentVersion, result.ExecutablePath, result.TargetVersion)
	} else {
		_, _ = fmt.Fprintf(stdout, "Install %s at %s? [y/N]: ", result.TargetVersion, result.ExecutablePath)
	}

	reader := bufio.NewReader(stdin)
	response, err := reader.ReadString('\n')
	if err != nil && len(response) == 0 {
		return false
	}

	answer := strings.ToLower(strings.TrimSpace(response))
	return answer == "y" || answer == "yes"
}

func renderUpdateCheckResult(stdout io.Writer, stderr io.Writer, result selfupdate.CheckResult, jsonOutput bool) int {
	if jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
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
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
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

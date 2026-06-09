package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/shiplah/refactorlah/internal/buildinfo"
)

type VersionCommand struct{}

func NewVersionCommand() *VersionCommand {
	return &VersionCommand{}
}

func (c *VersionCommand) Run(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("version", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var short bool
	var jsonOutput bool
	flags.BoolVar(&short, "short", false, "print only the version")
	flags.BoolVar(&jsonOutput, "json", false, "print JSON output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			WriteVersionUsage(stdout)
			return ExitSuccess
		}
		WriteCommandUsageError(stderr, err.Error())
		WriteVersionUsage(stderr)
		return ExitInvalidArguments
	}
	if flags.NArg() != 0 {
		WriteCommandUsageError(stderr, "version does not accept positional arguments")
		WriteVersionUsage(stderr)
		return ExitInvalidArguments
	}
	if short && jsonOutput {
		WriteCommandUsageError(stderr, "--short and --json cannot be used together")
		WriteVersionUsage(stderr)
		return ExitInvalidArguments
	}

	info := buildinfo.Current()
	switch {
	case jsonOutput:
		if err := writeJSONOutput(stdout, info); err != nil {
			fmt.Fprintf(stderr, "error: write version output: %v\n", err)
			return ExitGeneralFailure
		}
	case short:
		_, _ = fmt.Fprintln(stdout, info.Version)
	default:
		_, _ = fmt.Fprintf(stdout, "refactorlah %s\n", info.Version)
		_, _ = fmt.Fprintf(stdout, "commit: %s\n", info.Commit)
		_, _ = fmt.Fprintf(stdout, "build_date: %s\n", info.BuildDate)
		_, _ = fmt.Fprintf(stdout, "distribution: %s\n", info.Distribution)
		_, _ = fmt.Fprintf(stdout, "target: %s\n", info.Target())
	}

	return ExitSuccess
}

func WriteVersionUsage(writer io.Writer) {
	_, _ = io.WriteString(writer, "Usage:\n")
	_, _ = io.WriteString(writer, "  refactorlah version [--short|--json]\n")
}

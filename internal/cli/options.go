package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"refactorlah/internal/planning"
)

var ErrHelpRequested = errors.New("help requested")

type UsageError struct {
	Message string
}

func (e *UsageError) Error() string {
	return e.Message
}

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

type Options struct {
	OldPath              string
	NewPath              string
	DryRun               bool
	Apply                bool
	RequireCleanWorktree bool
	MoveRequests         []planning.RequestedMove
	UseList              bool
	UseFile              string
	NoAdapters           bool
	NoValidation         bool
	RunTests             bool
	Format               OutputFormat
}

func ParseOptions(args []string, stderr io.Writer) (Options, error) {
	fs := flag.NewFlagSet("refactorlah", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var format string
	options := Options{}

	fs.BoolVar(&options.DryRun, "dry", false, "preview changes without writing files")
	fs.BoolVar(&options.RequireCleanWorktree, "require-clean-worktree", false, "require a clean git working tree before applying changes")
	fs.BoolVar(&options.UseList, "use-list", false, "accept repeated old-path,new-path move pairs as positional arguments")
	fs.StringVar(&options.UseFile, "use-file", "", "read old-path,new-path move pairs from a file")
	fs.BoolVar(&options.NoAdapters, "no-adapters", false, "disable semantic adapter analysis")
	fs.BoolVar(&options.NoValidation, "no-validation", false, "skip post-apply validation")
	fs.BoolVar(&options.RunTests, "run-tests", false, "run composer test during validation")
	fs.StringVar(&format, "format", string(FormatText), "output format: text or json")

	flagArgs, positionalArgs, err := splitFlagArgs(args)
	if err != nil {
		return Options{}, err
	}

	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Options{}, ErrHelpRequested
		}
		return Options{}, &UsageError{Message: err.Error()}
	}

	if options.UseList && options.UseFile != "" {
		return Options{}, &UsageError{Message: "--use-list and --use-file cannot be used together"}
	}

	if options.UseList {
		if len(positionalArgs) == 0 {
			return Options{}, &UsageError{Message: "expected at least one old-path,new-path pair after --use-list"}
		}
		options.MoveRequests = make([]planning.RequestedMove, 0, len(positionalArgs))
		for _, pair := range positionalArgs {
			request, err := parseMovePair(pair)
			if err != nil {
				return Options{}, &UsageError{Message: err.Error()}
			}
			options.MoveRequests = append(options.MoveRequests, request)
		}
	} else if options.UseFile != "" {
		if len(positionalArgs) != 0 {
			return Options{}, &UsageError{Message: "positional paths cannot be used with --use-file"}
		}
	} else {
		if len(positionalArgs) != 2 {
			return Options{}, &UsageError{Message: "expected <old-path> and <new-path>"}
		}
		options.OldPath = positionalArgs[0]
		options.NewPath = positionalArgs[1]
		options.MoveRequests = []planning.RequestedMove{{
			OldPath: options.OldPath,
			NewPath: options.NewPath,
		}}
	}

	options.Apply = !options.DryRun

	switch OutputFormat(format) {
	case FormatText, FormatJSON:
		options.Format = OutputFormat(format)
	default:
		return Options{}, &UsageError{Message: fmt.Sprintf("unsupported format %q", format)}
	}

	return options, nil
}

func WriteUsage(writer io.Writer) {
	WriteUsageHeader(writer)
	_, _ = fmt.Fprintln(writer, "")
	_, _ = fmt.Fprintln(writer, "By default, refactorlah applies file moves and replacements.")
}

func WriteUsageHeader(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "Usage:")
	_, _ = fmt.Fprintln(writer, "  refactorlah move <old-path> <new-path> [options]")
	_, _ = fmt.Fprintln(writer, "")
	_, _ = fmt.Fprintln(writer, "Examples:")
	_, _ = fmt.Fprintln(writer, "  refactorlah move app/Services/Billing app/Domain/Billing")
	_, _ = fmt.Fprintln(writer, "  refactorlah move templates/admin templates/backoffice --dry")
	_, _ = fmt.Fprintln(writer, "  refactorlah move 'adapters/php/src/Php/Rules/*ReplacementWorker.php' 'adapters/php/src/Php/Rules/*ReplacementRule.php' --dry")
	_, _ = fmt.Fprintln(writer, "  refactorlah move --use-list app/Foo.php,app/Bar.php tests/A.php,tests/B.php")
	_, _ = fmt.Fprintln(writer, "  refactorlah move --use-file moves.txt")
	_, _ = fmt.Fprintln(writer, "")
	_, _ = fmt.Fprintln(writer, "Options:")
	_, _ = fmt.Fprintln(writer, "  --dry                     Preview changes without writing files")
	_, _ = fmt.Fprintln(writer, "  --require-clean-worktree  Require a clean git working tree before applying changes")
	_, _ = fmt.Fprintln(writer, "  --use-list                Accept repeated old-path,new-path move pairs as positional arguments")
	_, _ = fmt.Fprintln(writer, "  --use-file                Read old-path,new-path move pairs from a file")
	_, _ = fmt.Fprintln(writer, "  --no-adapters             Disable semantic adapter analysis")
	_, _ = fmt.Fprintln(writer, "  --format=text             Human-readable output (default)")
	_, _ = fmt.Fprintln(writer, "  --format=json             Machine-readable output")
	_, _ = fmt.Fprintln(writer, "  --no-validation           Skip post-apply validation")
	_, _ = fmt.Fprintln(writer, "  --run-tests               Run composer test during validation")
	_, _ = fmt.Fprintln(writer, "  --help                    Show this help")
}

func splitFlagArgs(args []string) ([]string, []string, error) {
	flagArgs := make([]string, 0, len(args))
	positionals := make([]string, 0, 2)

	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == "--" {
			positionals = append(positionals, args[index+1:]...)
			break
		}

		if !strings.HasPrefix(argument, "-") || argument == "-" {
			positionals = append(positionals, argument)
			continue
		}

		flagArgs = append(flagArgs, argument)
		if argument == "--format" || argument == "--use-file" {
			if index+1 >= len(args) {
				return nil, nil, fmt.Errorf("%s requires a value", argument)
			}
			index++
			flagArgs = append(flagArgs, args[index])
		}
	}

	return flagArgs, positionals, nil
}

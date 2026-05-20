package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

type Options struct {
	OldPath      string
	NewPath      string
	DryRun       bool
	Apply        bool
	AllowDirty   bool
	AllowNoGit   bool
	NoAdapters   bool
	NoValidation bool
	RunTests     bool
	Format       OutputFormat
}

func ParseOptions(args []string, stderr io.Writer) (Options, error) {
	fs := flag.NewFlagSet("refactorlah", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var format string
	options := Options{}

	fs.BoolVar(&options.DryRun, "dry-run", false, "preview changes without writing files")
	fs.BoolVar(&options.Apply, "apply", false, "apply file moves and replacements")
	fs.BoolVar(&options.AllowDirty, "allow-dirty", false, "allow apply mode on a dirty git working tree")
	fs.BoolVar(&options.AllowNoGit, "allow-no-git", false, "allow apply mode outside git repositories")
	fs.BoolVar(&options.NoAdapters, "no-adapters", false, "disable semantic adapter analysis")
	fs.BoolVar(&options.NoValidation, "no-validation", false, "skip post-apply validation")
	fs.BoolVar(&options.RunTests, "run-tests", false, "run composer test during validation")
	fs.StringVar(&format, "format", string(FormatText), "output format: text or json")

	flagArgs, positionalArgs, err := splitFlagArgs(args)
	if err != nil {
		return Options{}, err
	}

	if err := fs.Parse(flagArgs); err != nil {
		return Options{}, err
	}

	if len(positionalArgs) != 2 {
		return Options{}, errors.New("expected <old-path> and <new-path>")
	}

	options.OldPath = positionalArgs[0]
	options.NewPath = positionalArgs[1]

	if !options.Apply && !options.DryRun {
		options.DryRun = true
	}

	if options.Apply && options.DryRun {
		return Options{}, errors.New("--dry-run and --apply cannot be used together")
	}

	switch OutputFormat(format) {
	case FormatText, FormatJSON:
		options.Format = OutputFormat(format)
	default:
		return Options{}, fmt.Errorf("unsupported format %q", format)
	}

	return options, nil
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
		if argument == "--format" {
			if index+1 >= len(args) {
				return nil, nil, errors.New("--format requires a value")
			}
			index++
			flagArgs = append(flagArgs, args[index])
		}
	}

	return flagArgs, positionals, nil
}

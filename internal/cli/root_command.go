package cli

import (
	"context"
	"fmt"
	"io"

	"refactorlah/internal/selfupdate"
)

type RootCommand struct {
	move    *Command
	version *VersionCommand
	update  *UpdateCommand
	helper  *selfupdate.ReplacementHelper
}

func NewRootCommand() *RootCommand {
	return &RootCommand{
		move:    NewCommand(),
		version: NewVersionCommand(),
		update:  NewUpdateCommand(),
		helper:  &selfupdate.ReplacementHelper{},
	}
}

func (c *RootCommand) Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		WriteRootUsageError(stderr, "expected command")
		return ExitInvalidArguments
	}

	switch args[0] {
	case "move":
		return c.move.Run(ctx, args[1:], stdout, stderr)
	case "version":
		return c.version.Run(args[1:], stdout, stderr)
	case "update":
		return c.update.Run(ctx, args[1:], stdout, stderr)
	case "--version":
		return c.version.Run([]string{"--short"}, stdout, stderr)
	case "help", "--help", "-h":
		WriteRootUsage(stdout)
		return ExitSuccess
	case selfupdate.ReplaceHelperCommand:
		return c.helper.Run(ctx, args[1:], stderr)
	default:
		WriteRootUsageError(stderr, fmt.Sprintf("unknown command %q", args[0]))
		return ExitInvalidArguments
	}
}

func WriteRootUsageError(writer io.Writer, message string) {
	WriteCommandUsageError(writer, message)
	WriteRootUsage(writer)
}

func WriteRootUsage(writer io.Writer) {
	_, _ = io.WriteString(writer, "Usage:\n")
	_, _ = io.WriteString(writer, "  refactorlah <command> [arguments]\n")
	WriteRootCommands(writer)
}

func WriteRootCommands(writer io.Writer) {
	_, _ = io.WriteString(writer, "\nCommands:\n")
	_, _ = io.WriteString(writer, "  move           Move files/directories and update deterministic language references\n")
	_, _ = io.WriteString(writer, "  version        Print version and build information\n")
	_, _ = io.WriteString(writer, "  update         Check for and install published binary updates\n")
}

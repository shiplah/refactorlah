package cli

import (
	"context"
	"fmt"
	"io"
)

type RootCommand struct {
	move *Command
}

func NewRootCommand() *RootCommand {
	return &RootCommand{
		move: NewCommand(),
	}
}

func (c *RootCommand) Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		WriteUsageError(stderr, "expected command")
		WriteRootCommands(stderr)
		return ExitInvalidArguments
	}

	switch args[0] {
	case "move":
		return c.move.Run(ctx, args[1:], stdout, stderr)
	case "help", "--help", "-h":
		WriteRootUsage(stdout)
		return ExitSuccess
	default:
		WriteUsageError(stderr, fmt.Sprintf("unknown command %q", args[0]))
		WriteRootCommands(stderr)
		return ExitInvalidArguments
	}
}

func WriteRootUsage(writer io.Writer) {
	_, _ = io.WriteString(writer, "Usage:\n")
	_, _ = io.WriteString(writer, "  refactorlah <command> [arguments]\n")
	WriteRootCommands(writer)
}

func WriteRootCommands(writer io.Writer) {
	_, _ = io.WriteString(writer, "\nCommands:\n")
	_, _ = io.WriteString(writer, "  move           Move files/directories and update deterministic references\n")
}

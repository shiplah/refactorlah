package cli

import (
	"context"
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
		return c.move.Run(ctx, args, stdout, stderr)
	}

	switch args[0] {
	case "move":
		return c.move.Run(ctx, args[1:], stdout, stderr)
	case "help", "--help", "-h":
		WriteRootUsage(stdout)
		return ExitSuccess
	default:
		return c.move.Run(ctx, args, stdout, stderr)
	}
}

func WriteRootUsage(writer io.Writer) {
	WriteUsageHeader(writer)
	_, _ = io.WriteString(writer, "\nCommands:\n")
	_, _ = io.WriteString(writer, "  move           Move files/directories and update deterministic references\n")
	_, _ = io.WriteString(writer, "\nShort form:\n")
	_, _ = io.WriteString(writer, "  refactorlah <old-path> <new-path> [options]\n")
}

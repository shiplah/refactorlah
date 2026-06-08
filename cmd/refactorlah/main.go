package main

import (
	"context"
	"os"

	"github.com/NickSdot/refactorlah/internal/cli"
)

func main() {
	command := cli.NewRootCommand()
	exitCode := command.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

package main

import (
	"context"
	"fmt"
	"os"

	"refactorlah/internal/cli"
)

func main() {
	command := cli.NewCommand()
	exitCode := command.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
	if exitCode != 0 {
		os.Exit(exitCode)
	}

	fmt.Fprintln(os.Stderr, "")
}

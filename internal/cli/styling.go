package cli

import (
	"fmt"
	"io"
	"os"
)

const (
	ansiRed   = "\x1b[31m"
	ansiReset = "\x1b[0m"
)

func WriteUsageError(writer io.Writer, message string) {
	label := "error:"
	if supportsANSI(writer) {
		label = ansiRed + label + ansiReset
	}

	_, _ = fmt.Fprintf(writer, "%s %s\n\n", label, message)
	WriteUsage(writer)
}

func WriteCommandUsageError(writer io.Writer, message string) {
	label := "error:"
	if supportsANSI(writer) {
		label = ansiRed + label + ansiReset
	}

	_, _ = fmt.Fprintf(writer, "%s %s\n\n", label, message)
}

func supportsANSI(writer io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

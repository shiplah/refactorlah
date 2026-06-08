package selfupdate

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const ReplaceHelperCommand = "__self-update-replace"

type ReplacementHelper struct{}

func (h *ReplacementHelper) Run(ctx context.Context, args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet(ReplaceHelperCommand, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var targetPath string
	var sourcePath string
	var cleanupDir string

	flags.StringVar(&targetPath, "target", "", "target executable path")
	flags.StringVar(&sourcePath, "source", "", "downloaded executable path")
	flags.StringVar(&cleanupDir, "cleanup-dir", "", "temporary directory to remove after replacement")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(stderr, "error: parse replacement helper flags: %v\n", err)
		return 1
	}
	if targetPath == "" || sourcePath == "" {
		fmt.Fprintln(stderr, "error: replacement helper requires --target and --source")
		return 1
	}

	if err := waitAndReplace(targetPath, sourcePath, cleanupDir); err != nil {
		fmt.Fprintf(stderr, "error: replace executable: %v\n", err)
		return 1
	}

	return 0
}

func waitAndReplace(targetPath string, sourcePath string, cleanupDir string) error {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := replaceExecutable(targetPath, sourcePath); err == nil {
			if cleanupDir != "" {
				if runtime.GOOS == "windows" {
					scheduleWindowsCleanup(cleanupDir)
				} else {
					_ = os.RemoveAll(cleanupDir)
				}
			}
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(200 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = errors.New("timed out waiting for executable to become replaceable")
	}
	return lastErr
}

func replaceExecutable(targetPath string, sourcePath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open replacement binary: %w", err)
	}
	defer func() {
		if source != nil {
			_ = source.Close()
		}
	}()

	mode := os.FileMode(0o755)
	if info, err := source.Stat(); err == nil && info.Mode().Perm() != 0 {
		mode = info.Mode().Perm()
	}

	tempPath := targetPath + ".tmp"
	_ = os.Remove(tempPath)
	if err := writeExecutableFile(tempPath, source, mode); err != nil {
		return err
	}
	if err := source.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("close replacement binary: %w", err)
	}
	source = nil

	if runtime.GOOS == "windows" {
		if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
			_ = os.Remove(tempPath)
			return fmt.Errorf("remove current executable: %w", err)
		}
	}

	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("rename replacement into place: %w", err)
	}

	_ = os.Remove(sourcePath)
	return nil
}

func scheduleWindowsCleanup(cleanupDir string) {
	command := exec.Command("cmd.exe", "/c", fmt.Sprintf(`ping 127.0.0.1 -n 2 > nul & rmdir /s /q "%s"`, cleanupDir))
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	_ = command.Start()
}

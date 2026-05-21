package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"refactorlah/internal/planning"
)

func ReadMoveFile(cwd string, path string) ([]planning.RequestedMove, error) {
	return readMoveFile(cwd, path)
}

func ExpandMoveList(inputs []string) ([]planning.RequestedMove, error) {
	requests := make([]planning.RequestedMove, 0, len(inputs))
	for _, input := range inputs {
		request, err := parseMovePair(input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func readMoveFile(cwd string, path string) ([]planning.RequestedMove, error) {
	absolute := path
	if !filepath.IsAbs(path) {
		absolute = filepath.Join(cwd, path)
	}

	file, err := os.Open(absolute)
	if err != nil {
		return nil, fmt.Errorf("read move file %q: %w", path, err)
	}
	defer file.Close()

	requests := []planning.RequestedMove{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		request, err := parseMovePair(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNumber, err)
		}
		requests = append(requests, request)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read move file %q: %w", path, err)
	}

	return requests, nil
}

func parseMovePair(value string) (planning.RequestedMove, error) {
	oldPath, newPath, ok := strings.Cut(value, ",")
	if !ok {
		return planning.RequestedMove{}, fmt.Errorf("invalid move %q; expected old-path,new-path", value)
	}

	oldPath = strings.TrimSpace(oldPath)
	newPath = strings.TrimSpace(newPath)
	if oldPath == "" || newPath == "" {
		return planning.RequestedMove{}, fmt.Errorf("invalid move %q; expected old-path,new-path", value)
	}

	return planning.RequestedMove{
		OldPath: oldPath,
		NewPath: newPath,
	}, nil
}

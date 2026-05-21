package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"refactorlah/internal/planning"
)

func expandWildcardRequests(projectRoot string, requests []planning.RequestedMove) ([]planning.RequestedMove, error) {
	expanded := make([]planning.RequestedMove, 0, len(requests))
	candidates, err := projectRelativeCandidates(projectRoot)
	if err != nil {
		return nil, err
	}

	for _, request := range requests {
		requestsForMove, err := expandWildcardRequest(request, candidates)
		if err != nil {
			return nil, err
		}
		expanded = append(expanded, requestsForMove...)
	}

	return expanded, nil
}

func expandWildcardRequest(request planning.RequestedMove, candidates []string) ([]planning.RequestedMove, error) {
	oldWildcardCount := strings.Count(request.OldPath, "*")
	newWildcardCount := strings.Count(request.NewPath, "*")

	if oldWildcardCount == 0 && newWildcardCount == 0 {
		return []planning.RequestedMove{request}, nil
	}
	if oldWildcardCount == 0 || newWildcardCount == 0 {
		return nil, fmt.Errorf("wildcard moves require * in both old-path and new-path")
	}
	if oldWildcardCount != newWildcardCount {
		return nil, fmt.Errorf("wildcard moves require the same number of * placeholders in old-path and new-path")
	}

	pattern, err := compileWildcardPattern(request.OldPath, oldWildcardCount)
	if err != nil {
		return nil, err
	}

	matches := make([]planning.RequestedMove, 0)
	for _, candidate := range candidates {
		captured := pattern.FindStringSubmatch(candidate)
		if captured == nil {
			continue
		}

		matches = append(matches, planning.RequestedMove{
			OldPath: candidate,
			NewPath: applyWildcardCaptures(request.NewPath, captured[1:]),
		})
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("wildcard path %q matched no files or directories", request.OldPath)
	}

	return matches, nil
}

func compileWildcardPattern(pattern string, wildcardCount int) (*regexp.Regexp, error) {
	quoted := regexp.QuoteMeta(pattern)
	quoted = strings.ReplaceAll(quoted, `\*`, `([^/]*)`)
	compiled, err := regexp.Compile("^" + quoted + "$")
	if err != nil {
		return nil, fmt.Errorf("compile wildcard path %q: %w", pattern, err)
	}
	if compiled.NumSubexp() != wildcardCount {
		return nil, fmt.Errorf("compile wildcard path %q: placeholder mismatch", pattern)
	}
	return compiled, nil
}

func applyWildcardCaptures(pattern string, captures []string) string {
	expanded := pattern
	for _, capture := range captures {
		expanded = strings.Replace(expanded, "*", capture, 1)
	}
	return expanded
}

func projectRelativeCandidates(projectRoot string) ([]string, error) {
	candidates := make([]string, 0)
	err := filepath.WalkDir(projectRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relative, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == "." {
			return nil
		}

		candidates = append(candidates, relative)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk project for wildcard expansion: %w", err)
	}

	sort.Strings(candidates)
	return candidates, nil
}

package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/reporting"
)

const (
	slowAnalysisThreshold = 5 * time.Second
	largeCandidateBytes   = 2 << 20
)

func reportSlowAnalysis(writer io.Writer, done <-chan struct{}, stopped chan<- struct{}, started time.Time, firstNotice time.Duration, repeat time.Duration) {
	defer close(stopped)
	if writer == nil {
		writer = io.Discard
	}

	timer := time.NewTimer(firstNotice)
	defer timer.Stop()
	select {
	case <-done:
		return
	case <-timer.C:
		_, _ = fmt.Fprintf(writer, "analysing references... %s elapsed\n", time.Since(started).Round(time.Second))
	}

	ticker := time.NewTicker(repeat)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(writer, "analysing references... %s elapsed\n", time.Since(started).Round(time.Second))
		}
	}
}

func performanceDiagnostics(elapsed time.Duration, stats scan.CandidateStats) []reporting.Diagnostic {
	if elapsed < slowAnalysisThreshold {
		return nil
	}

	diagnostic := reporting.Diagnostic{DurationMS: elapsed.Milliseconds(), CandidateFiles: stats.Files}
	duration := elapsed.Round(time.Second)
	switch {
	case stats.LargestBytes >= largeCandidateBytes:
		diagnostic.Code = "large-candidate-file"
		diagnostic.File = stats.LargestFile
		diagnostic.FileBytes = stats.LargestBytes
		diagnostic.Message = fmt.Sprintf("Reference analysis took %s; large candidate %s (%s) may slow parsing.", duration, stats.LargestFile, formatFileSize(stats.LargestBytes))
		if pattern, ok := phpstanCachePattern(stats.LargestFile); ok {
			diagnostic.Message += fmt.Sprintf(" This appears to be a PHPStan cache; if generated, add %q to .refactorlah.json exclude.", pattern)
		} else {
			diagnostic.Message += fmt.Sprintf(" If generated, add %q to .refactorlah.json exclude.", stats.LargestFile)
		}
	case stats.Files >= 1000:
		diagnostic.Code = "many-candidate-files"
		diagnostic.Message = fmt.Sprintf("Reference analysis took %s across %d candidate files. Review generated directories and .refactorlah.json exclusions.", duration, stats.Files)
	default:
		diagnostic.Code = "slow-reference-analysis"
		diagnostic.Message = fmt.Sprintf("Reference analysis took %s; no unusually large reference candidate was identified.", duration)
	}
	return []reporting.Diagnostic{diagnostic}
}

func phpstanCachePattern(file string) (string, bool) {
	parts := strings.Split(file, "/")
	for index, part := range parts {
		if !strings.EqualFold(part, "phpstan") {
			continue
		}
		for laterIndex := index + 1; laterIndex < len(parts); laterIndex++ {
			if strings.EqualFold(parts[laterIndex], "cache") ||
				laterIndex == len(parts)-1 && strings.EqualFold(parts[laterIndex], "resultCache.php") {
				return strings.Join(parts[:index+1], "/") + "/**", true
			}
		}
	}
	return "", false
}

func formatFileSize(bytes int64) string {
	return fmt.Sprintf("%.1f MiB", float64(bytes)/(1<<20))
}

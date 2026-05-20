package replacements

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/reporting"
)

var ErrConflict = errors.New("replacement conflict")

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(projectRoot string, replacements []adapterproto.Replacement) ([]reporting.ValidationResult, error) {
	grouped := map[string][]adapterproto.Replacement{}
	for _, replacement := range replacements {
		grouped[replacement.File] = append(grouped[replacement.File], replacement)
	}

	results := []reporting.ValidationResult{}
	for file, fileReplacements := range grouped {
		content, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
		if err != nil {
			return nil, fmt.Errorf("%w: read %s: %v", ErrConflict, file, err)
		}
		if err := validateFile(file, content, fileReplacements); err != nil {
			return nil, err
		}
		results = append(results, reporting.ValidationResult{
			Name:    "replacement validation",
			Status:  "ok",
			Message: fmt.Sprintf("%s: %d replacements validated", file, len(fileReplacements)),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Message < results[j].Message
	})
	return results, nil
}

func validateFile(file string, content []byte, replacements []adapterproto.Replacement) error {
	sorted := append([]adapterproto.Replacement(nil), replacements...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start == sorted[j].Start {
			return sorted[i].End < sorted[j].End
		}
		return sorted[i].Start < sorted[j].Start
	})

	lastEnd := -1
	for _, replacement := range sorted {
		if replacement.Start < 0 || replacement.End < replacement.Start || replacement.End > len(content) {
			return fmt.Errorf("%w: invalid replacement range in %s (%d,%d)", ErrConflict, file, replacement.Start, replacement.End)
		}
		if replacement.Start < lastEnd {
			return fmt.Errorf("%w: overlapping replacements in %s", ErrConflict, file)
		}
		lastEnd = replacement.End
	}

	return nil
}

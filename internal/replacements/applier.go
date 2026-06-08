package replacements

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
)

type Applier struct{}

func NewApplier() *Applier {
	return &Applier{}
}

func (a *Applier) Apply(projectRoot string, movedPaths map[string]string, replacements []adapterproto.Replacement) error {
	grouped := map[string][]adapterproto.Replacement{}
	for _, replacement := range Deduplicate(replacements) {
		targetFile := replacement.File
		if moved, ok := movedPaths[targetFile]; ok {
			targetFile = moved
		}
		grouped[targetFile] = append(grouped[targetFile], replacement)
	}

	for file, fileReplacements := range grouped {
		if err := a.applyFile(projectRoot, file, fileReplacements); err != nil {
			return err
		}
	}

	return nil
}

func (a *Applier) applyFile(projectRoot string, file string, replacements []adapterproto.Replacement) error {
	absolute := filepath.Join(projectRoot, filepath.FromSlash(file))
	content, err := os.ReadFile(absolute)
	if err != nil {
		return err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return err
	}

	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].Start > replacements[j].Start
	})

	updated := append([]byte(nil), content...)
	for _, replacement := range replacements {
		updated = append(updated[:replacement.Start], append([]byte(replacement.Replacement), updated[replacement.End:]...)...)
	}

	if err := os.WriteFile(absolute, updated, info.Mode()); err != nil {
		return fmt.Errorf("write %s: %w", file, err)
	}
	return nil
}

package files

import (
	"io/fs"
	"os"
	"path/filepath"
)

func CollectFiles(root string, relativePath string) ([]string, error) {
	absoluteRoot := filepath.Join(root, filepath.FromSlash(relativePath))
	collected := []string{}

	err := filepath.WalkDir(absoluteRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if entry.IsDir() {
			if rel != relativePath && IsIgnoredPath(rel) {
				return filepath.SkipDir
			}
			return nil
		}

		if IsIgnoredPath(rel) {
			return nil
		}

		if info, err := entry.Info(); err == nil && info.Mode().IsRegular() {
			collected = append(collected, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return collected, nil
}

func Exists(root string, relativePath string) (bool, os.FileInfo, error) {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, info, nil
}

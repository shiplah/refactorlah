package golang

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/planning"
)

type packageMoveMapping struct {
	OldPath      string
	NewPath      string
	OldImport    string
	NewImport    string
	OldPackage   string
	NewPackage   string
	FilePackages []filePackageMapping
}

type filePackageMapping struct {
	OldPath    string
	OldPackage string
	NewPackage string
}

func packageMoveMappings(projectRoot string, goRoot string, modulePath string, plan planning.MovePlan) ([]packageMoveMapping, []adapterproto.Warning, error) {
	groups := map[string]map[string]bool{}
	moveTargets := map[string]string{}

	for _, move := range plan.Moves {
		if filepath.Ext(move.OldPath) != ".go" || filepath.Ext(move.NewPath) != ".go" {
			continue
		}

		oldDirectory := path.Dir(move.OldPath)
		newDirectory := path.Dir(move.NewPath)
		if oldDirectory == newDirectory {
			continue
		}
		if groups[oldDirectory] == nil {
			groups[oldDirectory] = map[string]bool{}
		}
		groups[oldDirectory][newDirectory] = true
		moveTargets[move.OldPath] = move.NewPath
	}

	var mappings []packageMoveMapping
	var warnings []adapterproto.Warning
	for _, oldDirectory := range sortedGroupKeys(groups) {
		targets := sortedKeys(groups[oldDirectory])
		if len(targets) != 1 {
			warnings = append(warnings, adapterproto.Warning{
				File:    oldDirectory,
				Message: "Go package directory is split across multiple targets; semantic rewrites skipped.",
			})
			continue
		}
		newDirectory := targets[0]

		goFiles, err := goFilesInDirectory(projectRoot, oldDirectory)
		if err != nil {
			return nil, nil, err
		}
		if !allPackageFilesMoveToDirectory(goFiles, moveTargets, newDirectory) {
			warnings = append(warnings, adapterproto.Warning{
				File:    oldDirectory,
				Message: "Go file moved across package directories without moving the full package; semantic rewrites skipped.",
			})
			continue
		}

		filePackages, oldPackage, newPackage, err := packageNameMappings(projectRoot, oldDirectory, newDirectory, goFiles)
		if err != nil {
			warnings = append(warnings, adapterproto.Warning{
				File:    oldDirectory,
				Message: fmt.Sprintf("Go package names could not be analysed; semantic rewrites skipped: %v", err),
			})
			continue
		}
		oldImport, err := importPathForDirectory(projectRoot, goRoot, modulePath, oldDirectory)
		if err != nil {
			return nil, nil, err
		}
		newImport, err := importPathForDirectory(projectRoot, goRoot, modulePath, newDirectory)
		if err != nil {
			return nil, nil, err
		}

		mappings = append(mappings, packageMoveMapping{
			OldPath:      oldDirectory,
			NewPath:      newDirectory,
			OldImport:    oldImport,
			NewImport:    newImport,
			OldPackage:   oldPackage,
			NewPackage:   newPackage,
			FilePackages: filePackages,
		})
	}

	return mappings, warnings, nil
}

func importPathForDirectory(projectRoot string, goRoot string, modulePath string, directory string) (string, error) {
	absoluteDirectory := filepath.Join(projectRoot, filepath.FromSlash(directory))
	relativeDirectory, err := filepath.Rel(goRoot, absoluteDirectory)
	if err != nil {
		return "", err
	}
	if relativeDirectory == ".." || strings.HasPrefix(relativeDirectory, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("go package directory %s is outside module root %s", directory, goRoot)
	}

	relativeSlash := filepath.ToSlash(relativeDirectory)
	if relativeSlash == "." {
		return modulePath, nil
	}

	return modulePath + "/" + relativeSlash, nil
}

func goFilesInDirectory(projectRoot string, directory string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(projectRoot, filepath.FromSlash(directory)))
	if err != nil {
		return nil, err
	}

	var goFiles []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		goFiles = append(goFiles, path.Join(directory, entry.Name()))
	}
	sort.Strings(goFiles)
	return goFiles, nil
}

func allPackageFilesMoveToDirectory(goFiles []string, moveTargets map[string]string, newDirectory string) bool {
	for _, goFile := range goFiles {
		targetPath, ok := moveTargets[goFile]
		if !ok || path.Dir(targetPath) != newDirectory {
			return false
		}
	}
	return len(goFiles) > 0
}

func packageNameMappings(projectRoot string, oldDirectory string, newDirectory string, goFiles []string) ([]filePackageMapping, string, string, error) {
	oldBase, oldBaseOK := packageNameFromDirectory(oldDirectory)
	newBase, newBaseOK := packageNameFromDirectory(newDirectory)

	primaryPackages := map[string]bool{}
	filePackages := make([]filePackageMapping, 0, len(goFiles))
	for _, goFile := range goFiles {
		oldPackage, err := readGoPackageName(filepath.Join(projectRoot, filepath.FromSlash(goFile)))
		if err != nil {
			return nil, "", "", err
		}
		newPackage := oldPackage
		if oldBaseOK && newBaseOK {
			if oldPackage == oldBase {
				newPackage = newBase
			} else if oldPackage == oldBase+"_test" {
				newPackage = newBase + "_test"
			}
		}
		filePackages = append(filePackages, filePackageMapping{
			OldPath:    goFile,
			OldPackage: oldPackage,
			NewPackage: newPackage,
		})
		if !strings.HasSuffix(oldPackage, "_test") {
			primaryPackages[oldPackage] = true
		}
	}

	oldPackage, err := primaryPackageName(oldDirectory, primaryPackages, filePackages)
	if err != nil {
		return nil, "", "", err
	}
	newPackage := oldPackage
	if oldBaseOK && newBaseOK && oldPackage == oldBase {
		newPackage = newBase
	}

	return filePackages, oldPackage, newPackage, nil
}

func readGoPackageName(absolutePath string) (string, error) {
	source, err := os.ReadFile(absolutePath)
	if err != nil {
		return "", err
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, absolutePath, source, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	if file.Name == nil {
		return "", fmt.Errorf("missing package declaration")
	}
	return file.Name.Name, nil
}

func primaryPackageName(oldDirectory string, primaryPackages map[string]bool, filePackages []filePackageMapping) (string, error) {
	if len(primaryPackages) == 1 {
		for packageName := range primaryPackages {
			return packageName, nil
		}
	}
	if len(primaryPackages) > 1 {
		return "", fmt.Errorf("multiple primary package names in %s", oldDirectory)
	}
	for _, filePackage := range filePackages {
		return strings.TrimSuffix(filePackage.OldPackage, "_test"), nil
	}
	return "", fmt.Errorf("no Go package files in %s", oldDirectory)
}

func packageNameFromDirectory(directory string) (string, bool) {
	name := path.Base(strings.Trim(directory, "/"))
	if !isValidGoPackageName(name) {
		return "", false
	}
	return name, true
}

func isValidGoPackageName(name string) bool {
	if name == "" || token.Lookup(name).IsKeyword() {
		return false
	}
	for index, character := range name {
		if character == '_' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' {
			continue
		}
		if index > 0 && character >= '0' && character <= '9' {
			continue
		}
		return false
	}
	return true
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedGroupKeys(values map[string]map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

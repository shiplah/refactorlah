package golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/planning"
)

type symbolMoveMapping struct {
	OldPath    string
	NewPath    string
	OldImport  string
	NewImport  string
	OldPackage string
	NewPackage string
	OldSymbol  string
	NewSymbol  string
	Kind       string
}

type goSymbolDeclaration struct {
	Name string
	Kind string
}

type symbolRenameCandidate struct {
	OldName string
	NewName string
}

func symbolMoveMappings(projectRoot string, goRoot string, modulePath string, plan planning.MovePlan, packageMappings []packageMoveMapping) ([]symbolMoveMapping, []adapterproto.Warning, error) {
	packageMappingByMove := map[string]packageMoveMapping{}
	for _, mapping := range packageMappings {
		packageMappingByMove[mapping.OldPath+"\x00"+mapping.NewPath] = mapping
	}

	seenNewSymbols := map[string]string{}
	var mappings []symbolMoveMapping
	var warnings []adapterproto.Warning
	for _, move := range plan.Moves {
		if filepath.Ext(move.OldPath) != ".go" || filepath.Ext(move.NewPath) != ".go" {
			continue
		}

		oldBase := strings.TrimSuffix(path.Base(move.OldPath), ".go")
		newBase := strings.TrimSuffix(path.Base(move.NewPath), ".go")
		if oldBase == newBase {
			continue
		}

		packageInfo, ok, err := packageInfoForSymbolMove(projectRoot, goRoot, modulePath, move, packageMappingByMove)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}

		mapping, ok, warning, err := symbolMoveMappingForFile(projectRoot, move, oldBase, newBase, packageInfo)
		if err != nil {
			return nil, nil, err
		}
		if warning.Message != "" {
			warnings = append(warnings, warning)
		}
		if !ok {
			continue
		}

		key := mapping.OldImport + "\x00" + mapping.OldPackage + "\x00" + mapping.NewSymbol
		if previous, exists := seenNewSymbols[key]; exists {
			warnings = append(warnings, adapterproto.Warning{
				File:    move.OldPath,
				Message: fmt.Sprintf("Go symbol rename skipped because %s and %s both map to %s.", previous, move.OldPath, mapping.NewSymbol),
			})
			continue
		}
		seenNewSymbols[key] = move.OldPath
		mappings = append(mappings, mapping)
	}

	return mappings, warnings, nil
}

type symbolPackageInfo struct {
	OldDirectory string
	NewDirectory string
	OldImport    string
	NewImport    string
	OldPackage   string
	NewPackage   string
}

func packageInfoForSymbolMove(projectRoot string, goRoot string, modulePath string, move planning.FileMove, packageMappingByMove map[string]packageMoveMapping) (symbolPackageInfo, bool, error) {
	oldDirectory := path.Dir(move.OldPath)
	newDirectory := path.Dir(move.NewPath)

	if oldDirectory != newDirectory {
		packageMapping, ok := packageMappingByMove[oldDirectory+"\x00"+newDirectory]
		if !ok {
			return symbolPackageInfo{}, false, nil
		}
		filePackage, ok := filePackageMappingForPath(packageMapping, move.OldPath)
		if !ok {
			return symbolPackageInfo{}, false, nil
		}
		return symbolPackageInfo{
			OldDirectory: oldDirectory,
			NewDirectory: newDirectory,
			OldImport:    packageMapping.OldImport,
			NewImport:    packageMapping.NewImport,
			OldPackage:   filePackage.OldPackage,
			NewPackage:   filePackage.NewPackage,
		}, true, nil
	}

	goFiles, err := goFilesInDirectory(projectRoot, oldDirectory)
	if err != nil {
		return symbolPackageInfo{}, false, err
	}
	filePackages, oldPackage, newPackage, err := packageNameMappings(projectRoot, oldDirectory, newDirectory, goFiles)
	if err != nil {
		return symbolPackageInfo{}, false, nil
	}
	filePackage, ok := filePackageMappingForPath(packageMoveMapping{FilePackages: filePackages}, move.OldPath)
	if !ok {
		return symbolPackageInfo{}, false, nil
	}
	oldImport, err := importPathForDirectory(projectRoot, goRoot, modulePath, oldDirectory)
	if err != nil {
		return symbolPackageInfo{}, false, err
	}

	return symbolPackageInfo{
		OldDirectory: oldDirectory,
		NewDirectory: newDirectory,
		OldImport:    oldImport,
		NewImport:    oldImport,
		OldPackage:   oldPackageForFile(filePackage, oldPackage),
		NewPackage:   newPackage,
	}, true, nil
}

func oldPackageForFile(filePackage filePackageMapping, fallback string) string {
	if filePackage.OldPackage != "" {
		return filePackage.OldPackage
	}
	return fallback
}

func filePackageMappingForPath(mapping packageMoveMapping, oldPath string) (filePackageMapping, bool) {
	for _, filePackage := range mapping.FilePackages {
		if filePackage.OldPath == oldPath {
			return filePackage, true
		}
	}
	return filePackageMapping{}, false
}

func symbolMoveMappingForFile(projectRoot string, move planning.FileMove, oldBase string, newBase string, packageInfo symbolPackageInfo) (symbolMoveMapping, bool, adapterproto.Warning, error) {
	declarations, err := goSymbolDeclarationsInFile(filepath.Join(projectRoot, filepath.FromSlash(move.OldPath)))
	if err != nil {
		return symbolMoveMapping{}, false, adapterproto.Warning{}, err
	}

	candidates := symbolRenameCandidates(oldBase, newBase)
	var matches []symbolMoveMapping
	for _, declaration := range declarations {
		for _, candidate := range candidates {
			if declaration.Name != candidate.OldName || candidate.OldName == candidate.NewName {
				continue
			}
			matches = append(matches, symbolMoveMapping{
				OldPath:    move.OldPath,
				NewPath:    move.NewPath,
				OldImport:  packageInfo.OldImport,
				NewImport:  packageInfo.NewImport,
				OldPackage: packageInfo.OldPackage,
				NewPackage: packageInfo.NewPackage,
				OldSymbol:  candidate.OldName,
				NewSymbol:  candidate.NewName,
				Kind:       declaration.Kind,
			})
		}
	}
	if len(matches) == 0 {
		return symbolMoveMapping{}, false, adapterproto.Warning{}, nil
	}
	if len(matches) > 1 {
		return symbolMoveMapping{}, false, adapterproto.Warning{
			File:    move.OldPath,
			Message: "Go symbol rename skipped because multiple top-level declarations match the old file basename.",
		}, nil
	}

	if exists, err := packageContainsTopLevelSymbol(projectRoot, packageInfo.OldDirectory, packageInfo.OldPackage, matches[0].NewSymbol, matches[0].OldPath); err != nil {
		return symbolMoveMapping{}, false, adapterproto.Warning{}, err
	} else if exists {
		return symbolMoveMapping{}, false, adapterproto.Warning{
			File:    move.OldPath,
			Message: fmt.Sprintf("Go symbol rename skipped because %s already exists in package %s.", matches[0].NewSymbol, packageInfo.OldPackage),
		}, nil
	}

	return matches[0], true, adapterproto.Warning{}, nil
}

func goSymbolDeclarationsInFile(absolutePath string) ([]goSymbolDeclaration, error) {
	source, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, absolutePath, source, 0)
	if err != nil {
		return nil, err
	}

	return topLevelDeclarations(file), nil
}

func topLevelDeclarations(file *ast.File) []goSymbolDeclaration {
	var declarations []goSymbolDeclaration
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			if typed.Recv == nil && typed.Name != nil {
				declarations = append(declarations, goSymbolDeclaration{Name: typed.Name.Name, Kind: "function"})
			}
		case *ast.GenDecl:
			kind := ""
			switch typed.Tok {
			case token.TYPE:
				kind = "type"
			case token.CONST:
				kind = "const"
			case token.VAR:
				kind = "var"
			default:
				continue
			}
			for _, spec := range typed.Specs {
				switch specTyped := spec.(type) {
				case *ast.TypeSpec:
					if specTyped.Name != nil {
						declarations = append(declarations, goSymbolDeclaration{Name: specTyped.Name.Name, Kind: kind})
					}
				case *ast.ValueSpec:
					for _, name := range specTyped.Names {
						if name != nil {
							declarations = append(declarations, goSymbolDeclaration{Name: name.Name, Kind: kind})
						}
					}
				}
			}
		}
	}
	return declarations
}

func packageContainsTopLevelSymbol(projectRoot string, directory string, packageName string, symbolName string, exceptPath string) (bool, error) {
	goFiles, err := goFilesInDirectory(projectRoot, directory)
	if err != nil {
		return false, err
	}
	for _, goFile := range goFiles {
		declarations, err := goSymbolDeclarationsInFile(filepath.Join(projectRoot, filepath.FromSlash(goFile)))
		if err != nil {
			return false, err
		}
		filePackage, err := readGoPackageName(filepath.Join(projectRoot, filepath.FromSlash(goFile)))
		if err != nil {
			return false, err
		}
		if filePackage != packageName {
			continue
		}
		for _, declaration := range declarations {
			if declaration.Name == symbolName && goFile != exceptPath {
				return true, nil
			}
		}
	}
	return false, nil
}

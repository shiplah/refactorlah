package python

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/python/syntax"
	"github.com/shiplah/refactorlah/internal/planning"
)

type SourceRootResolver struct{}

func (r SourceRootResolver) Resolve(projectRoot string, moves []planning.FileMove) ([]string, error) {
	roots := map[string]bool{}

	if isDir(filepath.Join(projectRoot, "src")) {
		roots["src"] = true
	}

	hasTopLevelPackages, err := hasTopLevelPackage(projectRoot)
	if err != nil {
		return nil, err
	}
	if hasTopLevelPackages {
		roots["."] = true
	}

	for _, move := range moves {
		for _, movePath := range []string{move.OldPath, move.NewPath} {
			root, ok := nearestPackageParent(projectRoot, movePath)
			if ok {
				roots[root] = true
			}
		}
	}

	if len(roots) == 0 {
		if anyMoveStartsWith(moves, "src/") {
			roots["src"] = true
		} else {
			roots["."] = true
		}
	}

	result := make([]string, 0, len(roots))
	for root := range roots {
		result = append(result, root)
	}
	sort.Slice(result, func(left int, right int) bool {
		leftDepth := strings.Count(result[left], "/")
		rightDepth := strings.Count(result[right], "/")
		if leftDepth != rightDepth {
			return leftDepth > rightDepth
		}
		return len(result[left]) > len(result[right])
	})

	return result, nil
}

type ModuleMapping struct {
	OldPath   string
	NewPath   string
	OldModule string
	NewModule string
	OldLeaf   string
	NewLeaf   string
}

func (m ModuleMapping) ToSymbolMapping() adapterproto.SymbolMapping {
	return adapterproto.SymbolMapping{
		Kind:         "module",
		OldPath:      m.OldPath,
		NewPath:      m.NewPath,
		OldSymbol:    m.OldModule,
		NewSymbol:    m.NewModule,
		OldNamespace: syntax.Parent(m.OldModule),
		NewNamespace: syntax.Parent(m.NewModule),
		ShortName:    m.OldLeaf,
	}
}

type ModuleMapper struct {
	sourceRoots []string
}

func NewModuleMapper(sourceRoots []string) ModuleMapper {
	return ModuleMapper{sourceRoots: append([]string(nil), sourceRoots...)}
}

func (m ModuleMapper) Derive(moves []planning.FileMove) ([]ModuleMapping, []adapterproto.Warning) {
	var mappings []ModuleMapping
	var warnings []adapterproto.Warning
	for _, move := range moves {
		if !strings.HasSuffix(move.OldPath, ".py") || !strings.HasSuffix(move.NewPath, ".py") {
			continue
		}

		oldModule, oldOK := m.ModuleForPath(move.OldPath)
		newModule, newOK := m.ModuleForPath(move.NewPath)
		if !oldOK || !newOK {
			warnings = append(warnings, adapterproto.Warning{
				File:    move.OldPath,
				Message: "Python file is outside known source roots; semantic rewrites skipped",
			})
			continue
		}
		if oldModule == newModule {
			continue
		}

		mappings = append(mappings, ModuleMapping{
			OldPath:   move.OldPath,
			NewPath:   move.NewPath,
			OldModule: oldModule,
			NewModule: newModule,
			OldLeaf:   syntax.Leaf(oldModule),
			NewLeaf:   syntax.Leaf(newModule),
		})
	}

	return mappings, warnings
}

func (m ModuleMapper) ModuleForPath(relativePath string) (string, bool) {
	clean := strings.Trim(relativePath, "/")
	for _, root := range m.sourceRoots {
		prefix := ""
		if root != "." {
			prefix = strings.TrimRight(root, "/") + "/"
		}
		if prefix != "" && !strings.HasPrefix(clean, prefix) {
			continue
		}

		modulePath := clean
		if prefix != "" {
			modulePath = strings.TrimPrefix(clean, prefix)
		}
		if !strings.HasSuffix(modulePath, ".py") {
			continue
		}

		withoutExtension := strings.TrimSuffix(modulePath, ".py")
		parts := strings.Split(withoutExtension, "/")
		filtered := make([]string, 0, len(parts))
		for _, part := range parts {
			if part == "" || part == "__init__" {
				continue
			}
			filtered = append(filtered, part)
		}
		if len(filtered) == 0 {
			continue
		}
		return strings.Join(filtered, "."), true
	}

	return "", false
}

func (m ModuleMapper) PackageForPath(relativePath string) (string, bool) {
	module, ok := m.ModuleForPath(relativePath)
	if !ok {
		return "", false
	}
	if path.Base(relativePath) == "__init__.py" {
		return module, true
	}
	return syntax.Parent(module), true
}

func nearestPackageParent(projectRoot string, relativePath string) (string, bool) {
	current := path.Dir(strings.Trim(relativePath, "/"))
	packageRoot := current
	sawPackage := false

	for current != "." && current != "" && hasInitFile(projectRoot, current) {
		sawPackage = true
		packageRoot = current
		next := path.Dir(current)
		if next == current {
			break
		}
		current = next
	}

	if !sawPackage {
		return "", false
	}

	parent := path.Dir(packageRoot)
	if parent == "." {
		return ".", true
	}
	return parent, true
}

func hasTopLevelPackage(projectRoot string) (bool, error) {
	entries, err := os.ReadDir(projectRoot)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.IsDir() && hasInitFile(projectRoot, entry.Name()) {
			return true, nil
		}
	}
	return false, nil
}

func hasInitFile(projectRoot string, relativeDirectory string) bool {
	info, err := os.Stat(filepath.Join(projectRoot, filepath.FromSlash(relativeDirectory), "__init__.py"))
	return err == nil && !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func anyMoveStartsWith(moves []planning.FileMove, prefix string) bool {
	for _, move := range moves {
		if strings.HasPrefix(move.OldPath, prefix) || strings.HasPrefix(move.NewPath, prefix) {
			return true
		}
	}
	return false
}

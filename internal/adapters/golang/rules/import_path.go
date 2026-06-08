package rules

import (
	"go/parser"
	"go/token"
	"strconv"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
)

const ImportPathRuleName = "go.ImportPathRule"

type ImportPathMapping struct {
	OldImport string
	NewImport string
}

type ImportPathInput struct {
	File     string
	Mappings []ImportPathMapping
}

type ImportPathRule struct{}

func (r ImportPathRule) Collect(source []byte, input ImportPathInput) ([]adapterproto.Replacement, error) {
	if len(input.Mappings) == 0 {
		return nil, nil
	}

	mappings := map[string]string{}
	for _, mapping := range input.Mappings {
		if mapping.OldImport == "" || mapping.OldImport == mapping.NewImport {
			continue
		}
		mappings[mapping.OldImport] = mapping.NewImport
	}
	if len(mappings) == 0 {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, input.File, source, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var replacements []adapterproto.Replacement
	for _, importSpec := range file.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			continue
		}

		newImportPath, ok := mappings[importPath]
		if !ok {
			continue
		}

		tokenFile := fileSet.File(importSpec.Path.Pos())
		start := tokenFile.Offset(importSpec.Path.Pos()) + 1
		end := tokenFile.Offset(importSpec.Path.End()) - 1
		replacements = append(replacements, adapterproto.Replacement{
			File:        input.File,
			Start:       start,
			End:         end,
			Replacement: newImportPath,
			Reason:      "go-import-path",
			Rule:        ImportPathRuleName,
			Adapter:     "go",
		})
	}

	return replacements, nil
}

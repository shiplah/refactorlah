package planning

import "path/filepath"

type MovePlan struct {
	OldPath string     `json:"oldPath"`
	NewPath string     `json:"newPath"`
	IsDir   bool       `json:"isDir"`
	Moves   []FileMove `json:"moves"`
}

func (p MovePlan) TargetPaths() map[string]string {
	result := map[string]string{}
	for _, move := range p.Moves {
		result[move.OldPath] = move.NewPath
	}
	return result
}

func (p MovePlan) ContainsExtension(ext string) bool {
	for _, move := range p.Moves {
		if filepath.Ext(move.OldPath) == ext || filepath.Ext(move.NewPath) == ext {
			return true
		}
	}
	return false
}

package planning

type FileMove struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
	Tracked bool   `json:"tracked"`
	Mover   string `json:"mover"`
}

package replacements

type Replacement struct {
	File        string
	Start       int
	End         int
	Replacement string
	Reason      string
	Worker      string
	Adapter     string
}

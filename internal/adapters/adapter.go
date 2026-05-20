package adapters

type Config struct {
	Name    string
	Path    string
	Options RequestOptions
}

type Selection struct {
	Adapters []Config
}

func (s Selection) Names() []string {
	names := make([]string, 0, len(s.Adapters))
	for _, adapter := range s.Adapters {
		names = append(names, adapter.Name)
	}
	return names
}

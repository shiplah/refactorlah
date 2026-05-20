package reporting

import (
	"encoding/json"
	"io"
)

func RenderJSON(writer io.Writer, result Result) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

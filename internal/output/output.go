package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Printer struct {
	Format string
	Out    io.Writer
}

func (p Printer) Emit(v any, text func(io.Writer) error) error {
	switch strings.ToLower(p.Format) {
	case "", "json":
		enc := json.NewEncoder(p.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case "text":
		if text == nil {
			return fmt.Errorf("text format not available for this command")
		}
		return text(p.Out)
	default:
		return fmt.Errorf("unknown format %q (use json or text)", p.Format)
	}
}

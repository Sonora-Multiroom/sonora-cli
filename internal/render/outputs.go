// Package render formats hub data as YAML or JSON for CLI output.
package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sonora-cli/internal/hub"
)

// RenderYAML renders outputs as a small, fixed-shape YAML document (the
// default output format per constitution Principle V). Every field is
// always emitted explicitly — in particular available: false is never
// omitted, so it can't be mistaken for "available" (FR-005).
func RenderYAML(outputs []hub.Output) string {
	var b bytes.Buffer
	if len(outputs) == 0 {
		b.WriteString("# no outputs found\n")
		b.WriteString("outputs: []\n")
		return b.String()
	}
	b.WriteString("outputs:\n")
	for _, o := range outputs {
		fmt.Fprintf(&b, "  - outputId: %q\n", o.OutputID)
		fmt.Fprintf(&b, "    displayName: %q\n", o.DisplayName)
		fmt.Fprintf(&b, "    volume: %d\n", o.Volume)
		fmt.Fprintf(&b, "    muted: %t\n", o.Muted)
		fmt.Fprintf(&b, "    available: %t\n", o.Available)
		fmt.Fprintf(&b, "    enabled: %t\n", o.Enabled)
	}
	return b.String()
}

type jsonPayload struct {
	Outputs []hub.Output `json:"outputs"`
}

// RenderJSON renders outputs as strict, parseable JSON: {"outputs": [...]}
// (FR-007), with the same fields as RenderYAML.
func RenderJSON(outputs []hub.Output) string {
	if outputs == nil {
		outputs = []hub.Output{}
	}
	data, err := json.Marshal(jsonPayload{Outputs: outputs})
	if err != nil {
		// jsonPayload's fields are all plain scalars/structs — Marshal
		// cannot fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

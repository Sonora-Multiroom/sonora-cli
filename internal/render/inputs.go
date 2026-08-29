package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sonora-cli/internal/hub"
)

func writeCreatedAt(b *bytes.Buffer, indent string, createdAt *string) {
	if createdAt == nil {
		fmt.Fprintf(b, "%screatedAt: null\n", indent)
		return
	}
	fmt.Fprintf(b, "%screatedAt: %q\n", indent, *createdAt)
}

// RenderInputsYAML renders inputs as a small, fixed-shape YAML document (the
// default output format per constitution Principle V). Every field is
// always emitted explicitly, including a bare, unquoted createdAt: null for
// static inputs.
//
// Named RenderInputsYAML rather than RenderYAML (which would mirror
// outputs.go's list renderer) because Go has no function overloading by
// argument type — a same-named, same-package RenderYAML(outputs
// []hub.Output) already exists in internal/render/outputs.go.
func RenderInputsYAML(inputs []hub.Input) string {
	var b bytes.Buffer
	if len(inputs) == 0 {
		b.WriteString("# no inputs found\n")
		b.WriteString("inputs: []\n")
		return b.String()
	}
	b.WriteString("inputs:\n")
	for _, i := range inputs {
		fmt.Fprintf(&b, "  - inputId: %q\n", i.InputID)
		fmt.Fprintf(&b, "    displayName: %q\n", i.DisplayName)
		fmt.Fprintf(&b, "    uri: %q\n", i.URI)
		fmt.Fprintf(&b, "    source: %q\n", i.Source)
		fmt.Fprintf(&b, "    enabled: %t\n", i.Enabled)
		fmt.Fprintf(&b, "    autoRemove: %t\n", i.AutoRemove)
		fmt.Fprintf(&b, "    pauseable: %t\n", i.Pauseable)
		writeCreatedAt(&b, "    ", i.CreatedAt)
	}
	return b.String()
}

// RenderInputYAML renders a single input as a bare YAML record (no inputs:
// list wrapper), used by `inputs get` since exactly one input is ever
// returned. Every field is always emitted explicitly, including a bare,
// unquoted createdAt: null for a static input.
func RenderInputYAML(i hub.Input) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "inputId: %q\n", i.InputID)
	fmt.Fprintf(&b, "displayName: %q\n", i.DisplayName)
	fmt.Fprintf(&b, "uri: %q\n", i.URI)
	fmt.Fprintf(&b, "source: %q\n", i.Source)
	fmt.Fprintf(&b, "enabled: %t\n", i.Enabled)
	fmt.Fprintf(&b, "autoRemove: %t\n", i.AutoRemove)
	fmt.Fprintf(&b, "pauseable: %t\n", i.Pauseable)
	writeCreatedAt(&b, "", i.CreatedAt)
	return b.String()
}

type inputsJSONPayload struct {
	Inputs []hub.Input `json:"inputs"`
}

// RenderInputsJSON renders inputs as strict, parseable JSON:
// {"inputs": [...]} (FR-010), with the same fields as RenderInputsYAML. See
// RenderInputsYAML for why this isn't named RenderJSON.
func RenderInputsJSON(inputs []hub.Input) string {
	if inputs == nil {
		inputs = []hub.Input{}
	}
	data, err := json.Marshal(inputsJSONPayload{Inputs: inputs})
	if err != nil {
		// inputsJSONPayload's fields are all plain scalars/structs — Marshal
		// cannot fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// RenderInputJSON renders a single input as a strict JSON object (no list
// wrapper), used by `inputs get --json` since exactly one input is ever
// returned.
func RenderInputJSON(i hub.Input) string {
	data, err := json.Marshal(i)
	if err != nil {
		// hub.Input's fields are all plain scalars — Marshal cannot fail for
		// this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// inputDeletedPayload is the flat rendered view of an input-deletion result:
// deleteInput's 204 response has no body, so this is built directly from the
// inputID the caller acted on rather than a decoded hub.Input (mirroring
// routeDeletedPayload in internal/render/route.go).
type inputDeletedPayload struct {
	InputID string `json:"inputId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func toInputDeletedPayload(inputID, message string) inputDeletedPayload {
	return inputDeletedPayload{InputID: inputID, Status: "removed", Message: message}
}

// RenderInputDeletedYAML renders an input-deletion result as a bare YAML
// record, exposing exactly inputId, status, message in that order (mirroring
// RenderRouteDeletedYAML).
func RenderInputDeletedYAML(inputID, message string) string {
	payload := toInputDeletedPayload(inputID, message)
	var b bytes.Buffer
	fmt.Fprintf(&b, "inputId: %q\n", payload.InputID)
	fmt.Fprintf(&b, "status: %q\n", payload.Status)
	fmt.Fprintf(&b, "message: %q\n", payload.Message)
	return b.String()
}

// RenderInputDeletedJSON renders an input-deletion result as a strict JSON
// object, exposing exactly inputId, status, message.
func RenderInputDeletedJSON(inputID, message string) string {
	data, err := json.Marshal(toInputDeletedPayload(inputID, message))
	if err != nil {
		// inputDeletedPayload's fields are all plain strings — Marshal
		// cannot fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// Package render formats hub data as YAML or JSON for CLI output.
package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
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

// RenderOutputYAML renders a single output as a bare YAML record (no
// outputs: list wrapper), used by `outputs get` since exactly one output is
// ever returned. Every field is always emitted explicitly — in particular
// available: false is never omitted (FR-005).
func RenderOutputYAML(o hub.Output) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "outputId: %q\n", o.OutputID)
	fmt.Fprintf(&b, "displayName: %q\n", o.DisplayName)
	fmt.Fprintf(&b, "volume: %d\n", o.Volume)
	fmt.Fprintf(&b, "muted: %t\n", o.Muted)
	fmt.Fprintf(&b, "available: %t\n", o.Available)
	fmt.Fprintf(&b, "enabled: %t\n", o.Enabled)
	return b.String()
}

// RenderOutputVolumeYAML renders a single output-volume confirmation as a
// bare YAML record, used by `set outputs/<id> volume <n>`.
func RenderOutputVolumeYAML(ov hub.OutputVolume) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "outputId: %q\n", ov.OutputID)
	fmt.Fprintf(&b, "volume: %d\n", ov.Volume)
	fmt.Fprintf(&b, "updatedAt: %q\n", ov.UpdatedAt)
	return b.String()
}

// RenderOutputVolumeJSON renders a single output-volume confirmation as a
// strict JSON object, used by `set outputs/<id> volume <n> --json`.
func RenderOutputVolumeJSON(ov hub.OutputVolume) string {
	data, err := json.Marshal(ov)
	if err != nil {
		// hub.OutputVolume's fields are all plain scalars — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

type jsonPayload struct {
	Outputs []hub.Output `json:"outputs"`
}

// RenderOutputJSON renders a single output as a strict JSON object (no list
// wrapper), used by `outputs get --json` since exactly one output is ever
// returned.
func RenderOutputJSON(o hub.Output) string {
	data, err := json.Marshal(o)
	if err != nil {
		// hub.Output's fields are all plain scalars — Marshal cannot fail
		// for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// RenderOutputWithRoutesYAML renders a single output's record followed by
// the routes playing on it, used by `get outputs/<id> --include-routes`.
// Each route shows nine fields; `routes: []` stands for none, with no
// comment line, which in the middle of a record would read as a separate
// document.
func RenderOutputWithRoutesYAML(o hub.Output, routes []hub.RouteOnOutput) string {
	var b bytes.Buffer
	b.WriteString(RenderOutputYAML(o))
	if len(routes) == 0 {
		b.WriteString("routes: []\n")
		return b.String()
	}
	b.WriteString("routes:\n")
	for _, r := range routes {
		fmt.Fprintf(&b, "  - routeId: %q\n", r.RouteID)
		fmt.Fprintf(&b, "    inputId: %q\n", r.InputID)
		fmt.Fprintf(&b, "    targetId: %q\n", r.TargetID)
		fmt.Fprintf(&b, "    targetType: %q\n", r.TargetType)
		fmt.Fprintf(&b, "    status: %q\n", r.Status)
		fmt.Fprintf(&b, "    joinMode: %q\n", r.JoinMode)
		fmt.Fprintf(&b, "    duckState: %q\n", r.DuckState)
		fmt.Fprintf(&b, "    main: %t\n", r.Main)
		fmt.Fprintf(&b, "    paused: %t\n", r.Paused)
	}
	return b.String()
}

// outputWithRoutesPayload is one output's fields, promoted to the top
// level, followed by the routes playing on it.
type outputWithRoutesPayload struct {
	hub.Output
	Routes []hub.RouteOnOutput `json:"routes"`
}

// RenderOutputWithRoutesJSON renders a single output and the routes playing
// on it as one strict JSON object: the output's keys, then "routes" with
// every field of each route ([] for none). Used by `get outputs/<id>
// --include-routes --json`.
func RenderOutputWithRoutesJSON(o hub.Output, routes []hub.RouteOnOutput) string {
	if routes == nil {
		routes = []hub.RouteOnOutput{}
	}
	data, err := json.Marshal(outputWithRoutesPayload{Output: o, Routes: routes})
	if err != nil {
		// Every field is a plain scalar or a *string — Marshal cannot fail
		// for this input shape.
		panic(err)
	}
	return string(data) + "\n"
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

// RenderOutputMuteYAML renders a single output-mute confirmation as a bare
// YAML record, used by `mute outputs/<id>` and `unmute outputs/<id>`.
func RenderOutputMuteYAML(m hub.OutputMute) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "outputId: %q\n", m.OutputID)
	fmt.Fprintf(&b, "muted: %t\n", m.Muted)
	fmt.Fprintf(&b, "updatedAt: %q\n", m.UpdatedAt)
	return b.String()
}

// RenderOutputMuteJSON renders a single output-mute confirmation as a strict
// JSON object, used by `mute outputs/<id> --json` and `unmute outputs/<id> --json`.
func RenderOutputMuteJSON(m hub.OutputMute) string {
	data, err := json.Marshal(m)
	if err != nil {
		// hub.OutputMute's fields are all plain scalars — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

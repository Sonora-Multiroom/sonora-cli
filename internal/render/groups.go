package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sonora-cli/internal/hub"
)

func writeOutputIDs(b *bytes.Buffer, indent string, outputIDs []string) {
	if len(outputIDs) == 0 {
		fmt.Fprintf(b, "%soutputIds: []\n", indent)
		return
	}
	fmt.Fprintf(b, "%soutputIds:\n", indent)
	for _, id := range outputIDs {
		fmt.Fprintf(b, "%s  - %q\n", indent, id)
	}
}

// RenderGroupsYAML renders groups as a small, fixed-shape YAML document (the
// default output format per constitution Principle V), showing all five
// fields per group in the documented order (FR-004).
func RenderGroupsYAML(groups []hub.Group) string {
	var b bytes.Buffer
	if len(groups) == 0 {
		b.WriteString("# no groups found\n")
		b.WriteString("groups: []\n")
		return b.String()
	}
	b.WriteString("groups:\n")
	for _, g := range groups {
		fmt.Fprintf(&b, "  - groupId: %q\n", g.GroupID)
		fmt.Fprintf(&b, "    displayName: %q\n", g.DisplayName)
		writeOutputIDs(&b, "    ", g.OutputIDs)
		fmt.Fprintf(&b, "    muted: %t\n", g.Muted)
		fmt.Fprintf(&b, "    enabled: %t\n", g.Enabled)
	}
	return b.String()
}

// RenderGroupYAML renders a single group as a bare YAML record (no groups:
// list wrapper), used by `groups get` since exactly one group is ever
// returned. Every field is always emitted explicitly, including
// outputIds: [] when a group has no member outputs (FR-008).
func RenderGroupYAML(g hub.Group) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "groupId: %q\n", g.GroupID)
	fmt.Fprintf(&b, "displayName: %q\n", g.DisplayName)
	writeOutputIDs(&b, "", g.OutputIDs)
	fmt.Fprintf(&b, "muted: %t\n", g.Muted)
	fmt.Fprintf(&b, "enabled: %t\n", g.Enabled)
	return b.String()
}

type groupsJSONPayload struct {
	Groups []hub.Group `json:"groups"`
}

// RenderGroupsJSON renders groups as strict, parseable JSON:
// {"groups": [...]} (FR-011), with the same fields as RenderGroupsYAML.
func RenderGroupsJSON(groups []hub.Group) string {
	// A group with no member outputs decodes to a nil OutputIDs, which
	// Marshal would render as null. Normalize to [] so every group in the
	// list matches what RenderGroupsYAML and RenderGroupJSON emit, and so
	// consumers can index/len outputIds without a null check.
	normalized := make([]hub.Group, len(groups))
	for i, g := range groups {
		if g.OutputIDs == nil {
			g.OutputIDs = []string{}
		}
		normalized[i] = g
	}
	data, err := json.Marshal(groupsJSONPayload{Groups: normalized})
	if err != nil {
		// groupsJSONPayload's fields are all plain scalars/slices — Marshal
		// cannot fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// RenderGroupJSON renders a single group as a strict JSON object (no list
// wrapper), used by `groups get --json` since exactly one group is ever
// returned.
func RenderGroupJSON(g hub.Group) string {
	if g.OutputIDs == nil {
		g.OutputIDs = []string{}
	}
	data, err := json.Marshal(g)
	if err != nil {
		// hub.Group's fields are all plain scalars/slices — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// RenderGroupVolumeYAML renders a single group-volume confirmation as a
// bare YAML record, used by `set groups/<id> volume <n>`.
func RenderGroupVolumeYAML(gv hub.GroupVolume) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "groupId: %q\n", gv.GroupID)
	fmt.Fprintf(&b, "volume: %d\n", gv.Volume)
	fmt.Fprintf(&b, "updatedAt: %q\n", gv.UpdatedAt)
	return b.String()
}

// RenderGroupVolumeJSON renders a single group-volume confirmation as a
// strict JSON object, used by `set groups/<id> volume <n> --json`.
func RenderGroupVolumeJSON(gv hub.GroupVolume) string {
	data, err := json.Marshal(gv)
	if err != nil {
		// hub.GroupVolume's fields are all plain scalars — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

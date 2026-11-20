package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// writeChime writes a chime setting only when there is one: a missing chime
// means the output, group or input has no opinion, so a line would say
// nothing. "none" is a value and is shown.
func writeChime(b *bytes.Buffer, indent, key string, chime *string) {
	if chime != nil {
		fmt.Fprintf(b, "%s%s: %q\n", indent, key, *chime)
	}
}

// RenderChimesYAML renders the hub's chimes as a small, fixed-shape YAML
// document: each chime's id and durationMillis, or an empty list with a
// comment when the hub has none.
func RenderChimesYAML(chimes []hub.Chime) string {
	var b bytes.Buffer
	if len(chimes) == 0 {
		b.WriteString("# no chimes found\n")
		b.WriteString("chimes: []\n")
		return b.String()
	}
	b.WriteString("chimes:\n")
	for _, c := range chimes {
		fmt.Fprintf(&b, "  - id: %q\n", c.ID)
		fmt.Fprintf(&b, "    durationMillis: %d\n", c.DurationMillis)
	}
	return b.String()
}

type chimesJSONPayload struct {
	Chimes []hub.Chime `json:"chimes"`
}

// RenderChimesJSON renders the hub's chimes as strict JSON,
// {"chimes": [...]}, with the same fields as RenderChimesYAML.
func RenderChimesJSON(chimes []hub.Chime) string {
	if chimes == nil {
		chimes = []hub.Chime{}
	}
	data, err := json.Marshal(chimesJSONPayload{Chimes: chimes})
	if err != nil {
		// A string and an int per chime: Marshal cannot fail.
		panic(err)
	}
	return string(data) + "\n"
}

// RenderChimeYAML renders one chime as a bare YAML record (no chimes: list
// wrapper), used by `get chimes/<id>`.
func RenderChimeYAML(c hub.Chime) string {
	return fmt.Sprintf("id: %q\ndurationMillis: %d\n", c.ID, c.DurationMillis)
}

// RenderChimeJSON renders one chime as a strict JSON object, used by
// `get chimes/<id> --json`.
func RenderChimeJSON(c hub.Chime) string {
	data, err := json.Marshal(c)
	if err != nil {
		// A string and an int: Marshal cannot fail.
		panic(err)
	}
	return string(data) + "\n"
}

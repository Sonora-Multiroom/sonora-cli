package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sonora-cli/internal/hub"
)

// RenderMasterMuteYAML renders the master-mute singleton as a bare YAML
// record, used by `get master-mute`, `mute all`, and `unmute all`.
func RenderMasterMuteYAML(mm hub.MasterMute) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "muted: %t\n", mm.Muted)
	return b.String()
}

// RenderMasterMuteJSON renders the master-mute singleton as a strict JSON
// object, used by `get master-mute --json`, `mute all --json`, and `unmute
// all --json`.
func RenderMasterMuteJSON(mm hub.MasterMute) string {
	data, err := json.Marshal(mm)
	if err != nil {
		// hub.MasterMute's only field is a plain bool — Marshal cannot fail
		// for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

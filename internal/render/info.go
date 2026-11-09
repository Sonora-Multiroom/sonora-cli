package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// RenderInfoYAML renders hub info as a bare YAML record: product, version,
// apiVersion, then startedAt and uptimeSeconds when the hub sent them.
func RenderInfoYAML(i hub.Info) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "product: %q\n", i.Product)
	fmt.Fprintf(&b, "version: %q\n", i.Version)
	fmt.Fprintf(&b, "apiVersion: %q\n", i.APIVersion)
	if i.StartedAt != "" {
		fmt.Fprintf(&b, "startedAt: %q\n", i.StartedAt)
	}
	if i.UptimeSeconds != nil {
		fmt.Fprintf(&b, "uptimeSeconds: %d\n", *i.UptimeSeconds)
	}
	return b.String()
}

// RenderInfoJSON renders hub info as a strict JSON object with the hub's
// property names; startedAt and uptimeSeconds are left out when absent.
func RenderInfoJSON(i hub.Info) string {
	data, err := json.Marshal(i)
	if err != nil {
		// hub.Info holds only strings and an *int64 — Marshal cannot fail.
		panic(err)
	}
	return string(data) + "\n"
}

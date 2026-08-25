package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sonora-cli/internal/hub"
)

func writeStartedAt(b *bytes.Buffer, indent string, startedAt *string) {
	if startedAt == nil {
		fmt.Fprintf(b, "%sstartedAt: null\n", indent)
		return
	}
	fmt.Fprintf(b, "%sstartedAt: %q\n", indent, *startedAt)
}

// RenderRoutesYAML renders routes as a small, fixed-shape YAML document (the
// default output format per constitution Principle V), showing only the
// five list-view fields per route (FR-004).
func RenderRoutesYAML(routes []hub.Route) string {
	var b bytes.Buffer
	if len(routes) == 0 {
		b.WriteString("# no routes found\n")
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
	}
	return b.String()
}

// RenderRouteYAML renders a single route as a bare YAML record (no routes:
// list wrapper), used by `routes get` since exactly one route is ever
// returned. All ten fields are always emitted explicitly, including a bare,
// unquoted startedAt: null before playback starts (FR-007).
func RenderRouteYAML(r hub.Route) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "routeId: %q\n", r.RouteID)
	fmt.Fprintf(&b, "inputId: %q\n", r.InputID)
	fmt.Fprintf(&b, "targetId: %q\n", r.TargetID)
	fmt.Fprintf(&b, "targetType: %q\n", r.TargetType)
	fmt.Fprintf(&b, "status: %q\n", r.Status)
	fmt.Fprintf(&b, "createdAt: %q\n", r.CreatedAt)
	writeStartedAt(&b, "", r.StartedAt)
	fmt.Fprintf(&b, "transferable: %t\n", r.Transferable)
	fmt.Fprintf(&b, "pauseable: %t\n", r.Pauseable)
	fmt.Fprintf(&b, "paused: %t\n", r.Paused)
	return b.String()
}

type routeListView struct {
	RouteID    string `json:"routeId"`
	InputID    string `json:"inputId"`
	TargetID   string `json:"targetId"`
	TargetType string `json:"targetType"`
	Status     string `json:"status"`
}

type routesJSONPayload struct {
	Routes []routeListView `json:"routes"`
}

// RenderRoutesJSON renders routes as strict, parseable JSON:
// {"routes": [...]} (FR-009), containing only the five list-view fields per
// route, mirroring RenderRoutesYAML's field set.
func RenderRoutesJSON(routes []hub.Route) string {
	views := make([]routeListView, len(routes))
	for i, r := range routes {
		views[i] = routeListView{
			RouteID:    r.RouteID,
			InputID:    r.InputID,
			TargetID:   r.TargetID,
			TargetType: r.TargetType,
			Status:     r.Status,
		}
	}
	data, err := json.Marshal(routesJSONPayload{Routes: views})
	if err != nil {
		// routesJSONPayload's fields are all plain scalars — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// RenderRouteJSON renders a single route as a strict JSON object (no list
// wrapper) with all ten fields, used by `routes get --json` since exactly
// one route is ever returned.
func RenderRouteJSON(r hub.Route) string {
	data, err := json.Marshal(r)
	if err != nil {
		// hub.Route's fields are all plain scalars — Marshal cannot fail for
		// this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

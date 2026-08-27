package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sonora-cli/internal/hub"
)

// routeCreatedPayload is the flat rendered view of a route-creation result:
// only routeId/status/message (FR-005) — the full hub.Route is deliberately
// not rendered (data-model.md's "RoutingResult" section).
type routeCreatedPayload struct {
	RouteID string `json:"routeId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func toRouteCreatedPayload(r hub.Route, message string) routeCreatedPayload {
	return routeCreatedPayload{RouteID: r.RouteID, Status: r.Status, Message: message}
}

// RenderRouteCreatedYAML renders a route-creation result as a bare YAML
// record, exposing exactly routeId, status, message in that order (FR-005).
func RenderRouteCreatedYAML(r hub.Route, message string) string {
	payload := toRouteCreatedPayload(r, message)
	var b bytes.Buffer
	fmt.Fprintf(&b, "routeId: %q\n", payload.RouteID)
	fmt.Fprintf(&b, "status: %q\n", payload.Status)
	fmt.Fprintf(&b, "message: %q\n", payload.Message)
	return b.String()
}

// RenderRouteCreatedJSON renders a route-creation result as a strict JSON
// object, exposing exactly routeId, status, message (FR-005).
func RenderRouteCreatedJSON(r hub.Route, message string) string {
	data, err := json.Marshal(toRouteCreatedPayload(r, message))
	if err != nil {
		// routeCreatedPayload's fields are all plain strings — Marshal
		// cannot fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

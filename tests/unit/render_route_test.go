package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func TestRenderRouteCreatedYAML_ExposesExactlyThreeFields(t *testing.T) {
	r := hub.Route{RouteID: "route_abc123", InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT", Status: "STARTING"}
	got := render.RenderRouteCreatedYAML(r, "Routed inputs/spotify-1 to outputs/office-speaker.")

	for _, field := range []string{"routeId", "status", "message"} {
		if !strings.Contains(got, field) {
			t.Errorf("expected field %q in YAML output, got:\n%s", field, got)
		}
	}
	if strings.Contains(got, "inputId") || strings.Contains(got, "targetId") || strings.Contains(got, "targetType") {
		t.Errorf("expected only routeId/status/message, got:\n%s", got)
	}
}

func TestRenderRouteCreatedJSON_RoundTrips(t *testing.T) {
	r := hub.Route{RouteID: "route_abc123", InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT", Status: "STARTING"}
	message := "Routed inputs/spotify-1 to outputs/office-speaker."
	got := render.RenderRouteCreatedJSON(r, message)

	var decoded struct {
		RouteID string `json:"routeId"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if decoded.RouteID != r.RouteID || decoded.Status != r.Status || decoded.Message != message {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

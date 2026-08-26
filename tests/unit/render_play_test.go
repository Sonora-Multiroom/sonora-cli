package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func testPlaybackResponse() hub.PlaybackResponse {
	return hub.PlaybackResponse{
		InputID: "playback_1782345678",
		Route:   hub.Route{RouteID: "route_abc123", Status: "STARTING"},
		Message: "Playback started: Radio Stream → office-speaker",
	}
}

func TestRenderPlaybackYAML_RendersExactFieldsInOrder(t *testing.T) {
	got := render.RenderPlaybackYAML(testPlaybackResponse())

	for _, want := range []string{
		`inputId: "playback_1782345678"`,
		`routeId: "route_abc123"`,
		`status: "STARTING"`,
		`message: "Playback started: Radio Stream → office-speaker"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("YAML output missing %q; got:\n%s", want, got)
		}
	}

	inputIdx := strings.Index(got, "inputId:")
	routeIdx := strings.Index(got, "routeId:")
	statusIdx := strings.Index(got, "status:")
	messageIdx := strings.Index(got, "message:")
	if !(inputIdx < routeIdx && routeIdx < statusIdx && statusIdx < messageIdx) {
		t.Errorf("expected documented field order inputId, routeId, status, message, got:\n%s", got)
	}
}

func TestRenderPlaybackJSON_RoundTripsWithExactFields(t *testing.T) {
	got := render.RenderPlaybackJSON(testPlaybackResponse())

	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if len(decoded) != 4 {
		t.Errorf("expected exactly 4 fields, got: %+v", decoded)
	}
	if decoded["inputId"] != "playback_1782345678" || decoded["routeId"] != "route_abc123" ||
		decoded["status"] != "STARTING" || decoded["message"] != "Playback started: Radio Stream → office-speaker" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

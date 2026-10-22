package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
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

func TestRenderPlaybackYAML_DisplayNameSecondWhenSet(t *testing.T) {
	p := testPlaybackResponse()
	p.DisplayName = "Flickermood — Forss"
	got := render.RenderPlaybackYAML(p)

	if !strings.Contains(got, `displayName: "Flickermood — Forss"`) {
		t.Errorf("YAML output missing unescaped displayName; got:\n%s", got)
	}
	var keys []string
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		key, _, _ := strings.Cut(line, ":")
		keys = append(keys, key)
	}
	want := []string{"inputId", "displayName", "routeId", "status", "message"}
	if strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Errorf("field order = %v, want %v; got:\n%s", keys, want, got)
	}
}

func TestRenderPlaybackJSON_DisplayNameWhenSet(t *testing.T) {
	p := testPlaybackResponse()
	p.DisplayName = "Flickermood — Forss"
	got := render.RenderPlaybackJSON(p)

	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if len(decoded) != 5 {
		t.Errorf("expected exactly 5 fields, got: %+v", decoded)
	}
	if decoded["displayName"] != "Flickermood — Forss" {
		t.Errorf("displayName = %v, want %q", decoded["displayName"], "Flickermood — Forss")
	}
	if !strings.Contains(got, "—") {
		t.Errorf("expected the em dash unescaped in raw JSON, got: %s", got)
	}
}

func TestRenderPlayback_CyrillicDisplayNameUnescaped(t *testing.T) {
	p := testPlaybackResponse()
	p.DisplayName = "Кино — Группа крови"

	if got := render.RenderPlaybackYAML(p); !strings.Contains(got, `displayName: "Кино — Группа крови"`) {
		t.Errorf("YAML output does not carry the name verbatim; got:\n%s", got)
	}
	if got := render.RenderPlaybackJSON(p); !strings.Contains(got, `"displayName":"Кино — Группа крови"`) {
		t.Errorf("JSON output does not carry the name verbatim; got: %s", got)
	}
}

func TestRenderPlayback_EmptyDisplayNameOmitted(t *testing.T) {
	p := testPlaybackResponse()
	p.DisplayName = ""

	wantYAML := "inputId: \"playback_1782345678\"\n" +
		"routeId: \"route_abc123\"\n" +
		"status: \"STARTING\"\n" +
		"message: \"Playback started: Radio Stream → office-speaker\"\n"
	if got := render.RenderPlaybackYAML(p); got != wantYAML {
		t.Errorf("YAML output changed;\ngot:\n%s\nwant:\n%s", got, wantYAML)
	}

	wantJSON := `{"inputId":"playback_1782345678","routeId":"route_abc123","status":"STARTING","message":"Playback started: Radio Stream → office-speaker"}` + "\n"
	got := render.RenderPlaybackJSON(p)
	if got != wantJSON {
		t.Errorf("JSON output changed;\ngot:  %s\nwant: %s", got, wantJSON)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if len(decoded) != 4 {
		t.Errorf("expected exactly 4 fields, got: %+v", decoded)
	}
	if _, ok := decoded["displayName"]; ok {
		t.Errorf("expected displayName absent, got: %+v", decoded)
	}
}

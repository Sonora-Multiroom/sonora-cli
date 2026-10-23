package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func TestRenderOutputYAML_AllFieldsAsBareRecord(t *testing.T) {
	o := hub.Output{
		OutputID: "office-speaker", DisplayName: "Office Speaker",
		Volume: 75, Muted: false, Available: true, Enabled: true,
	}
	got := render.RenderOutputYAML(o)

	if strings.Contains(got, "outputs:") {
		t.Errorf("expected a bare record, not a list wrapper, got:\n%s", got)
	}
	for _, want := range []string{
		`outputId: "office-speaker"`, `displayName: "Office Speaker"`,
		"volume: 75", "muted: false", "available: true", "enabled: true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderOutputYAML_UnavailableShownExplicitly(t *testing.T) {
	o := hub.Output{
		OutputID: "garage-speaker", DisplayName: "Garage Speaker",
		Volume: 0, Muted: false, Available: false, Enabled: true,
	}
	got := render.RenderOutputYAML(o)

	if !strings.Contains(got, "available: false") {
		t.Errorf("expected available: false to be shown explicitly, got:\n%s", got)
	}
}

func TestRenderOutputJSON_SingleObjectRoundTrips(t *testing.T) {
	o := hub.Output{
		OutputID: "office-speaker", DisplayName: "Office Speaker",
		Volume: 75, Muted: false, Available: true, Enabled: true,
	}
	got := render.RenderOutputJSON(o)

	if strings.Contains(got, `"outputs"`) || strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Errorf("expected a single object, not a list wrapper, got:\n%s", got)
	}

	var decoded hub.Output
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded != o {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, o)
	}
}

// outputWithRoutesFixture is one output and the two routes playing on it: a
// group route lowered under a ducking route, then that ducking route.
func outputWithRoutesFixture() (hub.Output, []hub.RouteOnOutput) {
	o := hub.Output{
		OutputID: "kitchen", DisplayName: "Kitchen",
		Volume: 30, Muted: false, Available: true, Enabled: true,
	}
	startedAt := "2026-09-30T08:00:01Z"
	routes := []hub.RouteOnOutput{
		{
			RouteID: "route-1", InputID: "radio", TargetID: "downstairs", TargetType: "OUTPUT_GROUP",
			Status: "ACTIVE", JoinMode: "REPLACE", CreatedAt: "2026-09-30T08:00:00Z", StartedAt: &startedAt,
			Transferable: true, Pauseable: true, Paused: false, DuckState: "LOWERED", Main: true,
		},
		{
			RouteID: "route-2", InputID: "doorbell", TargetID: "kitchen", TargetType: "SINGLE_OUTPUT",
			Status: "STARTING", JoinMode: "DUCK_OTHERS", CreatedAt: "2026-09-30T08:05:00Z",
			Transferable: false, Pauseable: false, Paused: false, DuckState: "FULL", Main: false,
		},
	}
	return o, routes
}

const outputRecordYAML = `outputId: "kitchen"
displayName: "Kitchen"
volume: 30
muted: false
available: true
enabled: true
`

func TestRenderOutputWithRoutesYAML_RoutesAfterRecord(t *testing.T) {
	o, routes := outputWithRoutesFixture()
	got := render.RenderOutputWithRoutesYAML(o, routes)

	want := outputRecordYAML + `routes:
  - routeId: "route-1"
    inputId: "radio"
    targetId: "downstairs"
    targetType: "OUTPUT_GROUP"
    status: "ACTIVE"
    joinMode: "REPLACE"
    duckState: "LOWERED"
    main: true
    paused: false
  - routeId: "route-2"
    inputId: "doorbell"
    targetId: "kitchen"
    targetType: "SINGLE_OUTPUT"
    status: "STARTING"
    joinMode: "DUCK_OTHERS"
    duckState: "FULL"
    main: false
    paused: false
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderOutputWithRoutesYAML_NoRoutes(t *testing.T) {
	o, _ := outputWithRoutesFixture()
	got := render.RenderOutputWithRoutesYAML(o, nil)

	if want := outputRecordYAML + "routes: []\n"; got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderOutputWithRoutesJSON_FlatObjectWithRoutes(t *testing.T) {
	o, routes := outputWithRoutesFixture()
	got := render.RenderOutputWithRoutesJSON(o, routes)

	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	for _, k := range []string{"outputId", "displayName", "volume", "muted", "available", "enabled", "routes"} {
		if _, ok := decoded[k]; !ok {
			t.Errorf("expected key %q, got: %s", k, got)
		}
	}
	list, ok := decoded["routes"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("expected routes to be a two-entry array, got: %s", got)
	}
	for i, entry := range list {
		m := entry.(map[string]any)
		for _, k := range []string{
			"routeId", "inputId", "targetId", "targetType", "status", "joinMode", "createdAt",
			"startedAt", "transferable", "pauseable", "paused", "duckState", "main",
		} {
			if _, ok := m[k]; !ok {
				t.Errorf("route %d: expected key %q, got: %v", i, k, m)
			}
		}
		if len(m) != 13 {
			t.Errorf("route %d: got %d keys, want 13: %v", i, len(m), m)
		}
	}
}

func TestRenderOutputWithRoutesJSON_NoRoutesIsEmptyArray(t *testing.T) {
	o, _ := outputWithRoutesFixture()
	got := render.RenderOutputWithRoutesJSON(o, nil)

	if !strings.Contains(got, `"routes":[]`) {
		t.Errorf("expected \"routes\":[], got: %s", got)
	}
}

func TestRenderOutputWithRoutes_UnknownDuckStateShownAsSent(t *testing.T) {
	o, routes := outputWithRoutesFixture()
	routes[0].DuckState = "MUFFLED"

	if got := render.RenderOutputWithRoutesYAML(o, routes); !strings.Contains(got, `duckState: "MUFFLED"`) {
		t.Errorf("expected MUFFLED in YAML, got:\n%s", got)
	}
	if got := render.RenderOutputWithRoutesJSON(o, routes); !strings.Contains(got, `"duckState":"MUFFLED"`) {
		t.Errorf("expected MUFFLED in JSON, got:\n%s", got)
	}
}

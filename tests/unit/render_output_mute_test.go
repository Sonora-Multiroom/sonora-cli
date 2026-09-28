package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func TestRenderOutputMuteYAML_AllFieldsAsBareRecord(t *testing.T) {
	m := hub.OutputMute{OutputID: "office-speaker", Muted: true, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderOutputMuteYAML(m)

	for _, want := range []string{
		`outputId: "office-speaker"`, "muted: true", `updatedAt: "2026-06-22T14:30:00Z"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderOutputMuteJSON_SingleObjectRoundTrips(t *testing.T) {
	m := hub.OutputMute{OutputID: "office-speaker", Muted: false, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderOutputMuteJSON(m)

	var raw map[string]any
	if err := json.Unmarshal([]byte(got), &raw); err != nil {
		t.Fatalf("output is not a JSON object: %v\ngot: %s", err, got)
	}
	if _, ok := raw["muted"]; !ok {
		t.Errorf("expected the muted key even when false, got:\n%s", got)
	}

	var decoded hub.OutputMute
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded != m {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, m)
	}
}

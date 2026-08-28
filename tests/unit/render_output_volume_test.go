package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func TestRenderOutputVolumeYAML_AllFieldsAsBareRecord(t *testing.T) {
	ov := hub.OutputVolume{OutputID: "office-speaker", Volume: 75, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderOutputVolumeYAML(ov)

	for _, want := range []string{
		`outputId: "office-speaker"`, "volume: 75", `updatedAt: "2026-06-22T14:30:00Z"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderOutputVolumeJSON_SingleObjectRoundTrips(t *testing.T) {
	ov := hub.OutputVolume{OutputID: "office-speaker", Volume: 75, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderOutputVolumeJSON(ov)

	if strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Errorf("expected a single object, not a list wrapper, got:\n%s", got)
	}

	var decoded hub.OutputVolume
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded != ov {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, ov)
	}
}

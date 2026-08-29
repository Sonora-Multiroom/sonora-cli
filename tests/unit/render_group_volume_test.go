package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func TestRenderGroupVolumeYAML_AllFieldsAsBareRecord(t *testing.T) {
	gv := hub.GroupVolume{GroupID: "downstairs", Volume: 75, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderGroupVolumeYAML(gv)

	for _, want := range []string{
		`groupId: "downstairs"`, "volume: 75", `updatedAt: "2026-06-22T14:30:00Z"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderGroupVolumeJSON_SingleObjectRoundTrips(t *testing.T) {
	gv := hub.GroupVolume{GroupID: "downstairs", Volume: 75, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderGroupVolumeJSON(gv)

	if strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Errorf("expected a single object, not a list wrapper, got:\n%s", got)
	}

	var decoded hub.GroupVolume
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded != gv {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, gv)
	}
}

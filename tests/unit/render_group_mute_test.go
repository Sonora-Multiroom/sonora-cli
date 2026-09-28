package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func TestRenderGroupMuteYAML_AllFieldsAsBareRecord(t *testing.T) {
	m := hub.GroupMute{GroupID: "living-room", Muted: true, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderGroupMuteYAML(m)

	for _, want := range []string{
		`groupId: "living-room"`, "muted: true", `updatedAt: "2026-06-22T14:30:00Z"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderGroupMuteJSON_SingleObjectRoundTrips(t *testing.T) {
	m := hub.GroupMute{GroupID: "living-room", Muted: false, UpdatedAt: "2026-06-22T14:30:00Z"}
	got := render.RenderGroupMuteJSON(m)

	var raw map[string]any
	if err := json.Unmarshal([]byte(got), &raw); err != nil {
		t.Fatalf("output is not a JSON object: %v\ngot: %s", err, got)
	}
	if _, ok := raw["muted"]; !ok {
		t.Errorf("expected the muted key even when false, got:\n%s", got)
	}

	var decoded hub.GroupMute
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded != m {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, m)
	}
}

package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func TestRenderGroupsYAML_ShowsSystemAfterEnabled(t *testing.T) {
	got := render.RenderGroupsYAML([]hub.Group{
		{GroupID: "broadcast", DisplayName: "Broadcast", Enabled: true, System: true},
		{GroupID: "mine", DisplayName: "Mine", Enabled: true},
	})
	if !strings.Contains(got, "system: true") || !strings.Contains(got, "system: false") {
		t.Fatalf("want system true and false, got:\n%s", got)
	}
	if strings.Index(got, "enabled:") > strings.Index(got, "system:") {
		t.Errorf("system must follow enabled, got:\n%s", got)
	}
}

func TestRenderGroupYAML_ShowsSystem(t *testing.T) {
	got := render.RenderGroupYAML(hub.Group{GroupID: "broadcast", DisplayName: "Broadcast", System: true})
	if !strings.Contains(got, "system: true") {
		t.Errorf("want system: true, got:\n%s", got)
	}
}

func TestRenderGroupJSON_ShowsSystem(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(render.RenderGroupJSON(hub.Group{GroupID: "b", DisplayName: "B", System: true})), &m); err != nil {
		t.Fatal(err)
	}
	if m["system"] != true {
		t.Errorf("want system true, got %v", m["system"])
	}
}

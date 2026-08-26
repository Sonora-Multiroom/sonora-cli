package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func TestRenderGroupsYAML_RendersAllFieldsInOrder(t *testing.T) {
	groups := []hub.Group{
		{GroupID: "living-room", DisplayName: "Living Room Speakers", OutputIDs: []string{"office-speaker", "bedroom-speaker"}, Muted: false, Enabled: true},
	}
	got := render.RenderGroupsYAML(groups)

	for _, want := range []string{
		`groupId: "living-room"`,
		`displayName: "Living Room Speakers"`,
		"outputIds:",
		"office-speaker",
		"bedroom-speaker",
		"muted: false",
		"enabled: true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("YAML output missing %q; got:\n%s", want, got)
		}
	}

	groupIdx := strings.Index(got, "groupId:")
	displayNameIdx := strings.Index(got, "displayName:")
	outputIdsIdx := strings.Index(got, "outputIds:")
	mutedIdx := strings.Index(got, "muted:")
	enabledIdx := strings.Index(got, "enabled:")
	if !(groupIdx < displayNameIdx && displayNameIdx < outputIdsIdx && outputIdsIdx < mutedIdx && mutedIdx < enabledIdx) {
		t.Errorf("expected documented field order, got:\n%s", got)
	}
}

func TestRenderGroupsYAML_EmptyOutputIDsRendersExplicitEmptyList(t *testing.T) {
	groups := []hub.Group{
		{GroupID: "empty-group", DisplayName: "Empty Group", OutputIDs: []string{}, Muted: false, Enabled: true},
	}
	got := render.RenderGroupsYAML(groups)

	if !strings.Contains(got, "outputIds: []") {
		t.Errorf("expected explicit outputIds: [], got:\n%s", got)
	}
}

func TestRenderGroupsYAML_ZeroGroupsIsUnambiguous(t *testing.T) {
	got := render.RenderGroupsYAML(nil)

	if !strings.Contains(got, "groups: []") {
		t.Errorf("expected an explicit empty groups list; got:\n%s", got)
	}
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "no groups") {
		t.Errorf("expected an unambiguous 'no groups' note; got:\n%s", got)
	}
}

func TestRenderGroupYAML_AllFieldsAsBareRecord(t *testing.T) {
	g := hub.Group{GroupID: "living-room", DisplayName: "Living Room Speakers", OutputIDs: []string{"office-speaker"}, Muted: false, Enabled: true}
	got := render.RenderGroupYAML(g)

	if strings.Contains(got, "groups:") {
		t.Errorf("expected a bare record, not a list wrapper, got:\n%s", got)
	}
	for _, want := range []string{
		`groupId: "living-room"`, `displayName: "Living Room Speakers"`,
		"outputIds:", "office-speaker", "muted: false", "enabled: true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderGroupYAML_EmptyOutputIDsShownExplicitly(t *testing.T) {
	g := hub.Group{GroupID: "empty-group", DisplayName: "Empty Group", OutputIDs: []string{}, Muted: false, Enabled: false}
	got := render.RenderGroupYAML(g)

	if !strings.Contains(got, "outputIds: []") {
		t.Errorf("expected outputIds: [] to be shown explicitly, got:\n%s", got)
	}
}

func TestRenderGroupsJSON_StrictlyParseable(t *testing.T) {
	groups := []hub.Group{
		{GroupID: "living-room", DisplayName: "Living Room Speakers", OutputIDs: []string{"office-speaker"}, Muted: false, Enabled: true},
	}
	got := render.RenderGroupsJSON(groups)

	var decoded struct {
		Groups []struct {
			GroupID     string   `json:"groupId"`
			DisplayName string   `json:"displayName"`
			OutputIDs   []string `json:"outputIds"`
			Muted       bool     `json:"muted"`
			Enabled     bool     `json:"enabled"`
		} `json:"groups"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if len(decoded.Groups) != 1 || decoded.Groups[0].GroupID != "living-room" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestRenderGroupsJSON_ZeroGroups(t *testing.T) {
	got := render.RenderGroupsJSON(nil)

	var decoded struct {
		Groups []hub.Group `json:"groups"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if decoded.Groups == nil || len(decoded.Groups) != 0 {
		t.Errorf("expected an explicit empty array, got: %s", got)
	}
}

func TestRenderGroupJSON_SingleObjectRoundTrips(t *testing.T) {
	g := hub.Group{GroupID: "living-room", DisplayName: "Living Room Speakers", OutputIDs: []string{"office-speaker"}, Muted: false, Enabled: true}
	got := render.RenderGroupJSON(g)

	if strings.Contains(got, `"groups"`) || strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Errorf("expected a single object, not a list wrapper, got:\n%s", got)
	}

	var decoded hub.Group
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded.GroupID != g.GroupID || decoded.DisplayName != g.DisplayName || len(decoded.OutputIDs) != len(g.OutputIDs) {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, g)
	}
}

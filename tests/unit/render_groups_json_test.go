package unit

import (
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

// A group with no member outputs decodes to a nil OutputIDs. Every renderer
// must show that as an empty list, never null, so consumers can index or
// count outputIds without a null check.
func TestRenderGroupsJSON_EmptyMembershipIsEmptyArray(t *testing.T) {
	groups := []hub.Group{
		{GroupID: "empty", DisplayName: "Empty", OutputIDs: nil},
		{GroupID: "full", DisplayName: "Full", OutputIDs: []string{"office-speaker"}},
	}

	got := render.RenderGroupsJSON(groups)

	if strings.Contains(got, "null") {
		t.Errorf("expected no null in the rendered list, got:\n%s", got)
	}
	want := `{"groupId":"empty","displayName":"Empty","outputIds":[],"muted":false,"enabled":false}`
	if !strings.Contains(got, want) {
		t.Errorf("expected %s, got:\n%s", want, got)
	}
}

func TestRenderGroupsJSON_MatchesSingleGroupRendering(t *testing.T) {
	g := hub.Group{GroupID: "empty", DisplayName: "Empty"}

	single := strings.TrimSpace(render.RenderGroupJSON(g))
	list := strings.TrimSpace(render.RenderGroupsJSON([]hub.Group{g}))

	if !strings.Contains(list, single) {
		t.Errorf("list rendering should embed the same object as the single rendering.\nsingle: %s\nlist:   %s", single, list)
	}
}

func TestRenderGroupsJSON_NoGroupsIsEmptyArray(t *testing.T) {
	got := strings.TrimSpace(render.RenderGroupsJSON(nil))

	if got != `{"groups":[]}` {
		t.Errorf(`expected {"groups":[]}, got %s`, got)
	}
}

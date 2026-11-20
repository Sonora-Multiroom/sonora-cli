package unit

import (
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func strp(s string) *string { return &s }

// A chime is shown only when the hub reports one; an output, group or input
// with none renders exactly as before HUB-040.

func TestRenderOutputs_ChimeShownOnlyWhenSet(t *testing.T) {
	plain := hub.Output{OutputID: "office", DisplayName: "Office", Volume: 30, Available: true, Enabled: true}
	withChime := plain
	withChime.Chime = strp("ding")
	withNone := plain
	withNone.Chime = strp("none")

	if got := render.RenderYAML([]hub.Output{plain}); strings.Contains(got, "chime") {
		t.Errorf("list: no chime set, but output mentions it:\n%s", got)
	}
	if got := render.RenderOutputYAML(plain); strings.Contains(got, "chime") {
		t.Errorf("record: no chime set, but output mentions it:\n%s", got)
	}
	if got := render.RenderYAML([]hub.Output{withChime}); !strings.Contains(got, "    enabled: true\n    chime: \"ding\"\n") {
		t.Errorf("list: want chime after enabled, indented as a list field, got:\n%s", got)
	}
	if got := render.RenderOutputYAML(withChime); !strings.HasSuffix(got, "enabled: true\nchime: \"ding\"\n") {
		t.Errorf("record: want chime last, got:\n%s", got)
	}
	if got := render.RenderOutputYAML(withNone); !strings.HasSuffix(got, "chime: \"none\"\n") {
		t.Errorf("record: none is a value and must be shown, got:\n%s", got)
	}
	if got := render.RenderOutputWithRoutesYAML(withChime, nil); !strings.Contains(got, "chime: \"ding\"\nroutes: []\n") {
		t.Errorf("with routes: want chime before routes, got:\n%s", got)
	}
}

func TestRenderOutputs_ChimeJSON(t *testing.T) {
	plain := hub.Output{OutputID: "office", DisplayName: "Office"}
	if got := render.RenderOutputJSON(plain); strings.Contains(got, "chime") {
		t.Errorf("no chime set, but JSON mentions it: %s", got)
	}
	if got := render.RenderJSON([]hub.Output{plain}); strings.Contains(got, "chime") {
		t.Errorf("no chime set, but JSON list mentions it: %s", got)
	}
	plain.Chime = strp("ding")
	if got := render.RenderOutputJSON(plain); !strings.Contains(got, `"chime":"ding"`) {
		t.Errorf("want chime in JSON, got: %s", got)
	}
	if got := render.RenderJSON([]hub.Output{plain}); !strings.Contains(got, `"chime":"ding"`) {
		t.Errorf("want chime in JSON list, got: %s", got)
	}
}

func TestRenderGroups_ChimeShownOnlyWhenSet(t *testing.T) {
	plain := hub.Group{GroupID: "downstairs", DisplayName: "Downstairs", OutputIDs: []string{"office"}, Enabled: true}
	withChime := plain
	withChime.Chime = strp("gong")

	if got := render.RenderGroupsYAML([]hub.Group{plain}); strings.Contains(got, "chime") {
		t.Errorf("list: no chime set, but output mentions it:\n%s", got)
	}
	if got := render.RenderGroupYAML(plain); strings.Contains(got, "chime") {
		t.Errorf("record: no chime set, but output mentions it:\n%s", got)
	}
	if got := render.RenderGroupsYAML([]hub.Group{withChime}); !strings.Contains(got, "    system: false\n    chime: \"gong\"\n") {
		t.Errorf("list: want chime after system, got:\n%s", got)
	}
	if got := render.RenderGroupYAML(withChime); !strings.HasSuffix(got, "system: false\nchime: \"gong\"\n") {
		t.Errorf("record: want chime last, got:\n%s", got)
	}
	if got := render.RenderGroupJSON(plain); strings.Contains(got, "chime") {
		t.Errorf("no chime set, but JSON mentions it: %s", got)
	}
	if got := render.RenderGroupsJSON([]hub.Group{withChime}); !strings.Contains(got, `"chime":"gong"`) {
		t.Errorf("want chime in JSON list, got: %s", got)
	}
	if got := render.RenderGroupJSON(withChime); !strings.Contains(got, `"chime":"gong"`) {
		t.Errorf("want chime in JSON record, got: %s", got)
	}
}

func TestRenderInputs_DefaultChimeShownOnlyWhenSet(t *testing.T) {
	src := "EPHEMERAL"
	plain := hub.Input{InputID: "doorbell", DisplayName: "Doorbell", URI: "u1", Source: src, Enabled: true}
	withChime := plain
	withChime.DefaultChime = strp("ding")

	if got := render.RenderInputsYAML([]hub.Input{plain}); strings.Contains(got, "hime") {
		t.Errorf("list: no chime set, but output mentions it:\n%s", got)
	}
	if got := render.RenderInputYAML(plain); strings.Contains(got, "hime") {
		t.Errorf("record: no chime set, but output mentions it:\n%s", got)
	}
	if got := render.RenderInputsYAML([]hub.Input{withChime}); !strings.Contains(got, "    createdAt: null\n    defaultChime: \"ding\"\n") {
		t.Errorf("list: want defaultChime after createdAt, got:\n%s", got)
	}
	if got := render.RenderInputYAML(withChime); !strings.HasSuffix(got, "defaultChime: \"ding\"\n") {
		t.Errorf("record: want defaultChime last, got:\n%s", got)
	}
	if got := render.RenderInputJSON(plain); strings.Contains(got, "hime") {
		t.Errorf("no chime set, but JSON mentions it: %s", got)
	}
	if got := render.RenderInputsJSON([]hub.Input{withChime}); !strings.Contains(got, `"defaultChime":"ding"`) {
		t.Errorf("want defaultChime in JSON list, got: %s", got)
	}
	if got := render.RenderInputJSON(withChime); !strings.Contains(got, `"defaultChime":"ding"`) {
		t.Errorf("want defaultChime in JSON record, got: %s", got)
	}
}

func TestRenderChimes(t *testing.T) {
	chimes := []hub.Chime{{ID: "ding", DurationMillis: 1800}, {ID: "gong", DurationMillis: 3200}}
	want := "chimes:\n  - id: \"ding\"\n    durationMillis: 1800\n  - id: \"gong\"\n    durationMillis: 3200\n"
	if got := render.RenderChimesYAML(chimes); got != want {
		t.Errorf("YAML:\n%s\nwant:\n%s", got, want)
	}
	if got, want := render.RenderChimesYAML([]hub.Chime{}), "# no chimes found\nchimes: []\n"; got != want {
		t.Errorf("empty YAML:\n%s\nwant:\n%s", got, want)
	}
	if got, want := render.RenderChimesYAML(nil), "# no chimes found\nchimes: []\n"; got != want {
		t.Errorf("nil YAML:\n%s\nwant:\n%s", got, want)
	}
	wantJSON := `{"chimes":[{"id":"ding","durationMillis":1800},{"id":"gong","durationMillis":3200}]}` + "\n"
	if got := render.RenderChimesJSON(chimes); got != wantJSON {
		t.Errorf("JSON: %s want: %s", got, wantJSON)
	}
	if got, want := render.RenderChimesJSON(nil), `{"chimes":[]}`+"\n"; got != want {
		t.Errorf("nil JSON: %s want: %s", got, want)
	}
}

func TestRenderChime(t *testing.T) {
	c := hub.Chime{ID: "gong", DurationMillis: 3200}
	if got, want := render.RenderChimeYAML(c), "id: \"gong\"\ndurationMillis: 3200\n"; got != want {
		t.Errorf("YAML:\n%s\nwant:\n%s", got, want)
	}
	if got, want := render.RenderChimeJSON(c), `{"id":"gong","durationMillis":3200}`+"\n"; got != want {
		t.Errorf("JSON: %s want: %s", got, want)
	}
}

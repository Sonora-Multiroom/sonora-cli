package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func strPtr(s string) *string { return &s }

func TestRenderYAML_Inputs_RendersAllFieldsInOrder(t *testing.T) {
	inputs := []hub.Input{
		{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "https://stream.example.com/live.mp3", Source: "STATIC", Enabled: true, AutoRemove: false, Pauseable: true, CreatedAt: nil},
	}
	got := render.RenderInputsYAML(inputs)

	for _, want := range []string{
		`inputId: "spotify-1"`,
		`displayName: "Spotify Stream"`,
		`uri: "https://stream.example.com/live.mp3"`,
		`source: "STATIC"`,
		"enabled: true",
		"autoRemove: false",
		"pauseable: true",
		"createdAt: null",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("YAML output missing %q; got:\n%s", want, got)
		}
	}

	idIdx := strings.Index(got, "inputId:")
	nameIdx := strings.Index(got, "displayName:")
	uriIdx := strings.Index(got, "uri:")
	sourceIdx := strings.Index(got, "source:")
	enabledIdx := strings.Index(got, "enabled:")
	autoRemoveIdx := strings.Index(got, "autoRemove:")
	pauseableIdx := strings.Index(got, "pauseable:")
	createdAtIdx := strings.Index(got, "createdAt:")
	if !(idIdx < nameIdx && nameIdx < uriIdx && uriIdx < sourceIdx && sourceIdx < enabledIdx &&
		enabledIdx < autoRemoveIdx && autoRemoveIdx < pauseableIdx && pauseableIdx < createdAtIdx) {
		t.Errorf("expected documented field order, got:\n%s", got)
	}
}

func TestRenderYAML_Inputs_CreatedAtNullBareUnquoted(t *testing.T) {
	inputs := []hub.Input{
		{InputID: "spotify-1", DisplayName: "Spotify Stream", Source: "STATIC", CreatedAt: nil},
	}
	got := render.RenderInputsYAML(inputs)

	if !strings.Contains(got, "createdAt: null") {
		t.Errorf("expected bare unquoted 'createdAt: null', got:\n%s", got)
	}
	if strings.Contains(got, `createdAt: "null"`) {
		t.Errorf("createdAt: null must not be quoted, got:\n%s", got)
	}
}

func TestRenderYAML_Inputs_CreatedAtPopulatedIsQuoted(t *testing.T) {
	inputs := []hub.Input{
		{InputID: "line-in-1", DisplayName: "Line In", Source: "EPHEMERAL", CreatedAt: strPtr("2026-06-22T14:30:00Z")},
	}
	got := render.RenderInputsYAML(inputs)

	if !strings.Contains(got, `createdAt: "2026-06-22T14:30:00Z"`) {
		t.Errorf("expected quoted createdAt value, got:\n%s", got)
	}
}

func TestRenderYAML_Inputs_ZeroInputsIsUnambiguous(t *testing.T) {
	got := render.RenderInputsYAML(nil)

	if !strings.Contains(got, "inputs: []") {
		t.Errorf("expected an explicit empty inputs list; got:\n%s", got)
	}
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "no inputs") {
		t.Errorf("expected an unambiguous 'no inputs' note; got:\n%s", got)
	}
}

func TestRenderInputYAML_AllFieldsAsBareRecord(t *testing.T) {
	i := hub.Input{
		InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1",
		Source: "STATIC", Enabled: true, AutoRemove: false, Pauseable: true, CreatedAt: nil,
	}
	got := render.RenderInputYAML(i)

	if strings.Contains(got, "inputs:") {
		t.Errorf("expected a bare record, not a list wrapper, got:\n%s", got)
	}
	for _, want := range []string{
		`inputId: "spotify-1"`, `displayName: "Spotify Stream"`, `uri: "u1"`,
		`source: "STATIC"`, "enabled: true", "autoRemove: false", "pauseable: true", "createdAt: null",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderInputYAML_StaticShowsCreatedAtNullExplicitly(t *testing.T) {
	i := hub.Input{InputID: "spotify-1", DisplayName: "Spotify Stream", Source: "STATIC", CreatedAt: nil}
	got := render.RenderInputYAML(i)

	if !strings.Contains(got, "createdAt: null") {
		t.Errorf("expected createdAt: null shown explicitly, got:\n%s", got)
	}
}

func TestRenderJSON_Inputs_StrictlyParseable(t *testing.T) {
	inputs := []hub.Input{
		{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Source: "STATIC", Enabled: true, AutoRemove: false, Pauseable: true, CreatedAt: nil},
	}
	got := render.RenderInputsJSON(inputs)

	var decoded struct {
		Inputs []hub.Input `json:"inputs"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if len(decoded.Inputs) != 1 || decoded.Inputs[0].InputID != "spotify-1" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestRenderJSON_Inputs_ZeroInputs(t *testing.T) {
	got := render.RenderInputsJSON([]hub.Input(nil))

	var decoded struct {
		Inputs []hub.Input `json:"inputs"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if decoded.Inputs == nil || len(decoded.Inputs) != 0 {
		t.Errorf("expected an explicit empty array, got: %s", got)
	}
}

func TestRenderInputJSON_SingleObjectRoundTrips(t *testing.T) {
	i := hub.Input{
		InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1",
		Source: "STATIC", Enabled: true, AutoRemove: false, Pauseable: true, CreatedAt: strPtr("2026-06-22T14:30:00Z"),
	}
	got := render.RenderInputJSON(i)

	if strings.Contains(got, `"inputs"`) || strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Errorf("expected a single object, not a list wrapper, got:\n%s", got)
	}

	var decoded hub.Input
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded.InputID != i.InputID || decoded.DisplayName != i.DisplayName || *decoded.CreatedAt != *i.CreatedAt {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, i)
	}
}

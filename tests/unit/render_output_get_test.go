package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
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

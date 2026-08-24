package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func TestRenderYAML_RendersAllFields(t *testing.T) {
	outputs := []hub.Output{
		{OutputID: "office-speaker", DisplayName: "Office Speaker", Volume: 75, Muted: false, Available: true, Enabled: true},
	}
	got := render.RenderYAML(outputs)

	for _, want := range []string{
		`outputId: "office-speaker"`,
		`displayName: "Office Speaker"`,
		"volume: 75",
		"muted: false",
		"available: true",
		"enabled: true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("YAML output missing %q; got:\n%s", want, got)
		}
	}
}

func TestRenderYAML_ZeroOutputsIsUnambiguous(t *testing.T) {
	got := render.RenderYAML(nil)

	if !strings.Contains(got, "outputs: []") {
		t.Errorf("expected an explicit empty outputs list; got:\n%s", got)
	}
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "no outputs") {
		t.Errorf("expected an unambiguous 'no outputs' note; got:\n%s", got)
	}
}

func TestRenderYAML_UnavailableOutputIsExplicit(t *testing.T) {
	outputs := []hub.Output{
		{OutputID: "garage-speaker", DisplayName: "Garage Speaker", Volume: 40, Muted: true, Available: false, Enabled: true},
	}
	got := render.RenderYAML(outputs)

	if !strings.Contains(got, "available: false") {
		t.Errorf("expected an explicit 'available: false' line; got:\n%s", got)
	}
}

func TestRenderJSON_StrictlyParseable(t *testing.T) {
	outputs := []hub.Output{
		{OutputID: "office-speaker", DisplayName: "Office Speaker", Volume: 75, Muted: false, Available: true, Enabled: true},
	}
	got := render.RenderJSON(outputs)

	var decoded struct {
		Outputs []struct {
			OutputID    string `json:"outputId"`
			DisplayName string `json:"displayName"`
			Volume      int    `json:"volume"`
			Muted       bool   `json:"muted"`
			Available   bool   `json:"available"`
			Enabled     bool   `json:"enabled"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if len(decoded.Outputs) != 1 || decoded.Outputs[0].OutputID != "office-speaker" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestRenderJSON_ZeroOutputs(t *testing.T) {
	got := render.RenderJSON(nil)

	var decoded struct {
		Outputs []hub.Output `json:"outputs"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if decoded.Outputs == nil || len(decoded.Outputs) != 0 {
		t.Errorf("expected an explicit empty array, got: %s", got)
	}
	if !strings.Contains(got, `"outputs":[]`) && !strings.Contains(got, `"outputs": []`) {
		t.Errorf("expected {\"outputs\": []}, got: %s", got)
	}
}

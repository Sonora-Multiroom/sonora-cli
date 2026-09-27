package unit

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/api"
)

func TestAPISpec(t *testing.T) {
	want, err := os.ReadFile("../../api/openapi.json")
	if err != nil {
		t.Fatalf("reading api/openapi.json: %v", err)
	}

	if len(api.Spec) == 0 {
		t.Fatal("api.Spec is empty")
	}
	if !bytes.Equal(api.Spec, want) {
		t.Error("api.Spec is not byte-identical to api/openapi.json")
	}

	var doc map[string]any
	if err := json.Unmarshal(api.Spec, &doc); err != nil {
		t.Fatalf("api.Spec is not valid JSON: %v", err)
	}
	if _, ok := doc["openapi"].(string); !ok {
		t.Error("api.Spec has no string \"openapi\" field")
	}
}

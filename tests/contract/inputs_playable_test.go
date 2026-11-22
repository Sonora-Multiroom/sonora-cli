package contract

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Response shapes here mirror #/components/schemas/InputResponse in
// api/openapi.json, which since hub 0.1.35 (HUB-041) carries origin, playable
// and unplayableReason.

func TestGetInput_DecodesOriginAndPlayable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "radio-1", "displayName": "Radio", "uri": "radiobrowser://abc",
			"enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": false,
			"defaultJoinMode": nil, "defaultRouteVolume": nil, "defaultDuckingLevelDb": nil,
			"origin": "STORED", "playable": false,
			"unplayableReason": "no input handler or resolver supports scheme 'radiobrowser'",
		})
	}))
	defer srv.Close()

	input, err := hub.GetInput(context.Background(), hub.NewClient(), srv.URL, "radio-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Origin != "STORED" || input.Playable || input.UnplayableReason == nil ||
		*input.UnplayableReason != "no input handler or resolver supports scheme 'radiobrowser'" {
		t.Errorf("unexpected decoded input: %+v", input)
	}
}

func TestGetInput_PlayableHasNoReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "radio-1", "displayName": "Radio", "uri": "u",
			"enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": false,
			"defaultJoinMode": nil, "defaultRouteVolume": nil, "defaultDuckingLevelDb": nil,
			"origin": "CONFIGURATION", "playable": true, "unplayableReason": nil,
		})
	}))
	defer srv.Close()

	input, err := hub.GetInput(context.Background(), hub.NewClient(), srv.URL, "radio-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Origin != "CONFIGURATION" || !input.Playable || input.UnplayableReason != nil {
		t.Errorf("unexpected decoded input: %+v", input)
	}
}

// TestInput_MarshalsHub035Fields guards the shape a consumer that re-encodes
// an Input (sonora-mcp) publishes: the three fields are always present.
func TestInput_MarshalsHub035Fields(t *testing.T) {
	b, err := json.Marshal(hub.Input{InputID: "a", DisplayName: "A", Origin: "RUNTIME", Playable: true})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["origin"] != "RUNTIME" || got["playable"] != true {
		t.Errorf("origin/playable not marshalled: %s", b)
	}
	if v, ok := got["unplayableReason"]; !ok || v != nil {
		t.Errorf("unplayableReason should marshal as null: %s", b)
	}
}

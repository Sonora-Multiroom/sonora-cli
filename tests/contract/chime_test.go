package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Shapes here mirror the chime fields and GET /api/v2/chimes in api/openapi.json
// (HUB-040, constitution Principle II).

// jsonServer answers every request with status and body, and records the last
// request body (decoded) and the request line.
func jsonServer(t *testing.T, status int, response any) (*httptest.Server, *map[string]any, *string) {
	t.Helper()
	var gotBody map[string]any
	var gotLine string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLine = r.Method + " " + r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if s, ok := response.(string); ok {
			_, _ = w.Write([]byte(s))
			return
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(srv.Close)
	return srv, &gotBody, &gotLine
}

func TestPlayback_RequestBody_Chime(t *testing.T) {
	playResp := map[string]any{
		"inputId": "playback_1",
		"route":   map[string]any{"routeId": "route_1", "status": "STARTING"},
		"message": "Playback started",
	}
	base := hub.PlaybackRequest{URI: "https://stream.example.com/a.mp3", TargetID: "office", TargetType: "SINGLE_OUTPUT"}
	for name, tc := range map[string]struct {
		chime *string
		want  any // nil: key omitted
	}{
		"omitted when nil": {nil, nil},
		"an id":            {ptr("ding"), "ding"},
		"none":             {ptr(hub.ChimeNone), "none"},
	} {
		t.Run(name, func(t *testing.T) {
			srv, body, _ := jsonServer(t, 200, playResp)
			req := base
			req.Chime = tc.chime
			if _, err := hub.Playback(context.Background(), hub.NewClient(), srv.URL, req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, present := (*body)["chime"]
			if tc.want == nil && present {
				t.Errorf("chime present (%v), want omitted", got)
			}
			if tc.want != nil && got != tc.want {
				t.Errorf("chime = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCreateRoute_RequestBody_Chime(t *testing.T) {
	routeResp := map[string]any{
		"routeId": "route_1", "inputId": "doorbell", "targetId": "office", "targetType": "SINGLE_OUTPUT",
		"status": "STARTING", "createdAt": "2026-01-01T00:00:00Z", "startedAt": nil,
		"transferable": true, "pauseable": true, "paused": false,
	}
	base := hub.CreateRouteRequest{InputID: "doorbell", TargetID: "office", TargetType: "SINGLE_OUTPUT"}
	for name, tc := range map[string]struct {
		chime *string
		want  any
	}{
		"omitted when nil": {nil, nil},
		"an id":            {ptr("gong"), "gong"},
		"none":             {ptr(hub.ChimeNone), "none"},
	} {
		t.Run(name, func(t *testing.T) {
			srv, body, _ := jsonServer(t, 201, routeResp)
			req := base
			req.Chime = tc.chime
			if _, err := hub.CreateRoute(context.Background(), hub.NewClient(), srv.URL, req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, present := (*body)["chime"]
			if tc.want == nil && present {
				t.Errorf("chime present (%v), want omitted", got)
			}
			if tc.want != nil && got != tc.want {
				t.Errorf("chime = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCreateInput_RequestBody_DefaultChime(t *testing.T) {
	inputResp := map[string]any{
		"inputId": "doorbell", "displayName": "Doorbell", "uri": "u1", "enabled": true,
		"autoRemove": false, "source": "EPHEMERAL", "createdAt": "2026-01-01T00:00:00Z", "pauseable": false,
		"defaultChime": "ding",
	}
	srv, body, _ := jsonServer(t, 201, inputResp)
	req := hub.CreateInputRequest{InputID: "doorbell", DisplayName: "Doorbell", URI: "u1", DefaultChime: ptr("ding")}
	input, err := hub.CreateInput(context.Background(), hub.NewClient(), srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*body)["defaultChime"] != "ding" {
		t.Errorf("defaultChime = %v, want ding", (*body)["defaultChime"])
	}
	if input.DefaultChime == nil || *input.DefaultChime != "ding" {
		t.Errorf("decoded DefaultChime = %v, want ding", input.DefaultChime)
	}

	srv, body, _ = jsonServer(t, 201, inputResp)
	req.DefaultChime = nil
	if _, err := hub.CreateInput(context.Background(), hub.NewClient(), srv.URL, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := (*body)["defaultChime"]; present {
		t.Errorf("defaultChime present, want omitted when nil")
	}
}

// chimeCases are the three states a setting can be in on the wire.
var chimeCases = map[string]struct {
	wire any // the JSON value; absent when the key is left out
	omit bool
	want *string
}{
	"an id":  {wire: "ding", want: ptr("ding")},
	"none":   {wire: "none", want: ptr("none")},
	"null":   {wire: nil, want: nil},
	"absent": {omit: true, want: nil},
}

func TestGetOutput_DecodesChime(t *testing.T) {
	for name, tc := range chimeCases {
		t.Run(name, func(t *testing.T) {
			body := map[string]any{"outputId": "office", "displayName": "Office", "volume": 30, "muted": false, "available": true, "enabled": true}
			if !tc.omit {
				body["chime"] = tc.wire
			}
			srv, _, _ := jsonServer(t, 200, body)
			out, err := hub.GetOutput(context.Background(), hub.NewClient(), srv.URL, "office")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalPtr(out.Chime, tc.want) {
				t.Errorf("Chime = %v, want %v", out.Chime, tc.want)
			}
		})
	}
}

func TestGetGroup_DecodesChime(t *testing.T) {
	for name, tc := range chimeCases {
		t.Run(name, func(t *testing.T) {
			body := map[string]any{"groupId": "downstairs", "displayName": "Downstairs", "outputIds": []string{"office"}, "muted": false, "enabled": true}
			if !tc.omit {
				body["chime"] = tc.wire
			}
			srv, _, _ := jsonServer(t, 200, body)
			g, err := hub.GetGroup(context.Background(), hub.NewClient(), srv.URL, "downstairs")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalPtr(g.Chime, tc.want) {
				t.Errorf("Chime = %v, want %v", g.Chime, tc.want)
			}
		})
	}
}

func TestGetInput_DecodesDefaultChime(t *testing.T) {
	for name, tc := range chimeCases {
		t.Run(name, func(t *testing.T) {
			body := map[string]any{
				"inputId": "doorbell", "displayName": "Doorbell", "uri": "u1", "enabled": true,
				"autoRemove": false, "source": "EPHEMERAL", "createdAt": "2026-01-01T00:00:00Z", "pauseable": false,
			}
			if !tc.omit {
				body["defaultChime"] = tc.wire
			}
			srv, _, _ := jsonServer(t, 200, body)
			in, err := hub.GetInput(context.Background(), hub.NewClient(), srv.URL, "doorbell")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalPtr(in.DefaultChime, tc.want) {
				t.Errorf("DefaultChime = %v, want %v", in.DefaultChime, tc.want)
			}
		})
	}
}

// A nil setting is left out of the marshalled value, so a JSON view shows a
// chime only when there is one.
func TestChimeSettings_MarshalOmitsNil(t *testing.T) {
	for name, v := range map[string]any{"Output": hub.Output{}, "Group": hub.Group{}} {
		data, _ := json.Marshal(v)
		if _, present := keysAsSet(t, data)["chime"]; present {
			t.Errorf("%s: chime present in %s, want omitted when nil", name, data)
		}
	}
	data, _ := json.Marshal(hub.Input{})
	if _, present := keysAsSet(t, data)["defaultChime"]; present {
		t.Errorf("Input: defaultChime present in %s, want omitted when nil", data)
	}
	data, _ = json.Marshal(hub.Output{Chime: ptr("none")})
	if keysAsSet(t, data)["chime"] != "none" {
		t.Errorf("Output with chime none marshals to %s, want chime none", data)
	}
}

func keysAsSet(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("not a JSON object: %v: %s", err, data)
	}
	return m
}

func TestListChimes_RequestAndDecode(t *testing.T) {
	srv, _, line := jsonServer(t, 200, `{"chimes":[{"id":"ding","durationMillis":1800},{"id":"gong","durationMillis":3200}]}`)
	chimes, err := hub.ListChimes(context.Background(), hub.NewClient(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *line != "GET /api/v2/chimes" {
		t.Errorf("got %q, want GET /api/v2/chimes", *line)
	}
	if len(chimes) != 2 || chimes[0].ID != "ding" || chimes[0].DurationMillis != 1800 || chimes[1].ID != "gong" || chimes[1].DurationMillis != 3200 {
		t.Errorf("unexpected chimes: %+v", chimes)
	}
}

func TestListChimes_MarshalsWithSpecFieldNames(t *testing.T) {
	data, err := json.Marshal(hub.Chime{ID: "ding", DurationMillis: 1800})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got, want := keysOf(t, data), specSchemaProperties(t, "ChimeResponse"); !equalStrings(got, want) {
		t.Errorf("marshalled keys %v, want the spec's %v", got, want)
	}
}

func TestListChimes_EmptyListIsNotAnError(t *testing.T) {
	for name, body := range map[string]string{"empty array": `{"chimes":[]}`, "no chimes key": `{}`} {
		t.Run(name, func(t *testing.T) {
			srv, _, _ := jsonServer(t, 200, body)
			chimes, err := hub.ListChimes(context.Background(), hub.NewClient(), srv.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if chimes == nil || len(chimes) != 0 {
				t.Errorf("chimes = %#v, want a non-nil empty slice", chimes)
			}
		})
	}
}

func TestListChimes_BadBodyIsDecodeError(t *testing.T) {
	for name, body := range map[string]string{
		"not json":      `not json`,
		"not an object": `[1]`,
		"wrong type":    `{"chimes":"x"}`,
		"missing id":    `{"chimes":[{"durationMillis":10}]}`,
		"empty id":      `{"chimes":[{"id":"","durationMillis":10}]}`,
		"empty body":    ``,
	} {
		t.Run(name, func(t *testing.T) {
			srv, _, _ := jsonServer(t, 200, body)
			_, err := hub.ListChimes(context.Background(), hub.NewClient(), srv.URL)
			var decodeErr *hub.DecodeError
			if !errors.As(err, &decodeErr) {
				t.Fatalf("expected *hub.DecodeError, got %T: %v", err, err)
			}
		})
	}
}

func TestListChimes_NonSuccessIsStatusError(t *testing.T) {
	for _, status := range []int{404, 500} {
		srv, _, _ := jsonServer(t, status, ``)
		_, err := hub.ListChimes(context.Background(), hub.NewClient(), srv.URL)
		var statusErr *hub.StatusError
		if !errors.As(err, &statusErr) || statusErr.StatusCode != status {
			t.Errorf("status %d: expected *hub.StatusError with that code, got %T: %v", status, err, err)
		}
	}
}

func ptr[T any](v T) *T { return &v }

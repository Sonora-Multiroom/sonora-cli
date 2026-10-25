package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Request/response shapes here mirror #/components/schemas/CreateInputRequest,
// InputResponse, and ErrorResponse, and the createInput operation, in
// api/openapi.json (constitution Principle II).

func TestCreateInput_RequestBody_AlwaysIncludesRequiredFields(t *testing.T) {
	var gotBody map[string]any
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1",
			"enabled": true, "autoRemove": false, "source": "EPHEMERAL", "createdAt": "2026-01-01T00:00:00Z", "pauseable": true,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	enabled, autoRemove := true, false
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Enabled: &enabled, AutoRemove: &autoRemove}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["inputId"] != req.InputID || gotBody["displayName"] != req.DisplayName || gotBody["uri"] != req.URI {
		t.Errorf("expected inputId/displayName/uri always present, got body: %+v", gotBody)
	}
	if gotBody["enabled"] != true || gotBody["autoRemove"] != false {
		t.Errorf("expected enabled/autoRemove present when set, got body: %+v", gotBody)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry)", got)
	}
}

// TestCreateInput_RequestBody_OmitsUnsetEnabledAndAutoRemove verifies that
// leaving Enabled/AutoRemove nil omits them from the JSON body entirely,
// rather than sending an explicit false, so the hub's documented default of
// true (api/openapi.json CreateInputRequest.enabled/autoRemove) applies.
func TestCreateInput_RequestBody_OmitsUnsetEnabledAndAutoRemove(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1",
			"enabled": true, "autoRemove": true, "source": "EPHEMERAL", "createdAt": "2026-01-01T00:00:00Z", "pauseable": true,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1"}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := gotBody["enabled"]; ok {
		t.Errorf("expected enabled to be omitted when unset, got body: %+v", gotBody)
	}
	if _, ok := gotBody["autoRemove"]; ok {
		t.Errorf("expected autoRemove to be omitted when unset, got body: %+v", gotBody)
	}
	for _, k := range []string{"defaultJoinMode", "defaultRouteVolume", "defaultDuckingLevelDb"} {
		if _, ok := gotBody[k]; ok {
			t.Errorf("expected %s to be omitted when unset, got body: %+v", k, gotBody)
		}
	}
}

func TestCreateInput_RequestBody_IncludesDefaultJoinModeWhenSet(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1",
			"enabled": true, "autoRemove": true, "source": "STATIC", "createdAt": "2026-01-01T00:00:00Z", "pauseable": true,
			"defaultJoinMode": "DUCK_OTHERS",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	mode := hub.JoinModeDuckOthers
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", DefaultJoinMode: &mode}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["defaultJoinMode"] != "DUCK_OTHERS" {
		t.Errorf("expected defaultJoinMode DUCK_OTHERS, got body: %+v", gotBody)
	}
}

func TestCreateInput_Success_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/inputs" {
			t.Errorf("got path %q, want /api/v2/inputs", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("got method %q, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1",
			"enabled": true, "autoRemove": true, "source": "EPHEMERAL", "createdAt": "2026-01-01T00:00:00Z", "pauseable": true,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1"}
	resp, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.InputID != "spotify-1" || resp.DisplayName != "Spotify Stream" || resp.URI != "u1" {
		t.Errorf("unexpected decoded input: %+v", resp)
	}
	if resp.Source != "EPHEMERAL" || !resp.AutoRemove {
		t.Errorf("unexpected decoded input: %+v", resp)
	}
}

func testCreateInputErrorStatus(t *testing.T, status int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "urn:multiroom:error:validation-error", "title": "Validation Error", "detail": "inputId already exists",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1"}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatalf("expected an error for status %d, got nil", status)
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != status || apiErr.Title != "Validation Error" || apiErr.Detail != "inputId already exists" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestCreateInput_400_DecodesAsAPIError(t *testing.T) {
	testCreateInputErrorStatus(t, http.StatusBadRequest)
}
func TestCreateInput_409_DecodesAsAPIError(t *testing.T) {
	testCreateInputErrorStatus(t, http.StatusConflict)
}

func TestCreateInput_ErrorStatus_NonJSONBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1"}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a non-JSON error body, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError fallback, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusConflict {
		t.Errorf("got StatusCode %d, want 409", statusErr.StatusCode)
	}
}

func TestCreateInput_OtherErrorStatus_IsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1"}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("got StatusCode %d, want 500", statusErr.StatusCode)
	}
}

func TestCreateInput_MalformedBody_MissingInputID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"inputId":"","displayName":"Spotify Stream","uri":"u1","enabled":true,"autoRemove":false,"source":"EPHEMERAL"}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1"}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a malformed 201 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestCreateInput_RequestBody_IncludesRouteDefaultsWhenSet(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "doorbell", "displayName": "Doorbell", "uri": "u1",
			"enabled": true, "autoRemove": false, "source": "EPHEMERAL", "createdAt": "2026-01-01T00:00:00Z", "pauseable": false,
			"defaultRouteVolume": 90, "defaultDuckingLevelDb": -30.0,
		})
	}))
	defer srv.Close()

	volume, level := 90, -30.0
	req := hub.CreateInputRequest{InputID: "doorbell", DisplayName: "Doorbell", URI: "u1", DefaultRouteVolume: &volume, DefaultDuckingLevelDb: &level}
	input, err := hub.CreateInput(context.Background(), hub.NewClient(), srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["defaultRouteVolume"] != float64(90) || gotBody["defaultDuckingLevelDb"] != float64(-30) {
		t.Errorf("expected defaultRouteVolume 90 and defaultDuckingLevelDb -30, got body: %+v", gotBody)
	}
	if input.DefaultRouteVolume == nil || *input.DefaultRouteVolume != 90 || input.DefaultDuckingLevelDb == nil || *input.DefaultDuckingLevelDb != -30 {
		t.Errorf("unexpected decoded input: %+v", input)
	}
}

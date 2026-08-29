package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"sonora-cli/internal/hub"
)

// Request/response shapes here mirror #/components/schemas/CreateInputRequest,
// InputResponse, and ErrorResponse, and the createInput operation, in
// api/openapi.json (constitution Principle II).

func TestCreateInput_RequestBody_AlwaysIncludesAllFields(t *testing.T) {
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
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Enabled: true, AutoRemove: false}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["inputId"] != req.InputID || gotBody["displayName"] != req.DisplayName || gotBody["uri"] != req.URI {
		t.Errorf("expected inputId/displayName/uri always present, got body: %+v", gotBody)
	}
	if gotBody["enabled"] != true || gotBody["autoRemove"] != false {
		t.Errorf("expected enabled/autoRemove always present, got body: %+v", gotBody)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry)", got)
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
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Enabled: true, AutoRemove: true}
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
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Enabled: true}
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
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Enabled: true}
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
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Enabled: true}
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
	req := hub.CreateInputRequest{InputID: "spotify-1", DisplayName: "Spotify Stream", URI: "u1", Enabled: true}
	_, err := hub.CreateInput(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a malformed 201 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

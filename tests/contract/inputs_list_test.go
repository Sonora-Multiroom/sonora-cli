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

// Response shapes here mirror #/components/schemas/InputResponse and the
// listInputs operation in api/openapi.json (constitution Principle II).

func TestListInputs_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotQuery string
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("includeDisabled")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "https://stream.example.com/live.mp3", "enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
			{"inputId": "line-in-1", "displayName": "Line In", "uri": "line://1", "enabled": true, "autoRemove": true, "source": "EPHEMERAL", "createdAt": "2026-06-22T14:30:00Z", "pauseable": false},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	inputs, err := hub.ListInputs(context.Background(), client, srv.URL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/inputs" {
		t.Errorf("got path %q, want /api/v2/inputs", gotPath)
	}
	if gotQuery != "false" {
		t.Errorf("got includeDisabled=%q, want %q", gotQuery, "false")
	}
	if len(inputs) != 2 {
		t.Fatalf("got %d inputs, want 2", len(inputs))
	}
	if inputs[0].InputID != "spotify-1" || inputs[0].DisplayName != "Spotify Stream" ||
		inputs[0].Source != "STATIC" || inputs[0].CreatedAt != nil {
		t.Errorf("unexpected decoded static input: %+v", inputs[0])
	}
	if inputs[1].Source != "EPHEMERAL" || inputs[1].CreatedAt == nil || *inputs[1].CreatedAt != "2026-06-22T14:30:00Z" {
		t.Errorf("unexpected decoded ephemeral input: %+v", inputs[1])
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-013)", got)
	}
}

func TestListInputs_IncludeDisabledTrueSetsQueryParam(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("includeDisabled")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer srv.Close()

	client := hub.NewClient()
	if _, err := hub.ListInputs(context.Background(), client, srv.URL, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "true" {
		t.Errorf("got includeDisabled=%q, want %q", gotQuery, "true")
	}
}

func TestListInputs_MixedEnabledDisabledDecodesCorrectly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
			{"inputId": "line-in-1", "displayName": "Line In", "uri": "u2", "enabled": false, "autoRemove": true, "source": "EPHEMERAL", "createdAt": "2026-06-22T14:30:00Z", "pauseable": false},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	inputs, err := hub.ListInputs(context.Background(), client, srv.URL, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 2 || inputs[0].Enabled != true || inputs[1].Enabled != false {
		t.Errorf("unexpected decoded inputs: %+v", inputs)
	}
}

func TestListInputs_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListInputs(context.Background(), client, srv.URL, false)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestListInputs_MalformedBodyRejected_MissingInputID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"displayName":"Spotify Stream","uri":"u1","enabled":true,"autoRemove":false,"source":"STATIC","createdAt":null,"pauseable":true}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListInputs(context.Background(), client, srv.URL, false)
	if err == nil {
		t.Fatal("expected an error for a malformed body (missing inputId), got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestListInputs_MalformedBodyRejected_UnrecognizedSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"inputId":"spotify-1","displayName":"Spotify Stream","uri":"u1","enabled":true,"autoRemove":false,"source":"BOGUS","createdAt":null,"pauseable":true}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListInputs(context.Background(), client, srv.URL, false)
	if err == nil {
		t.Fatal("expected an error for a malformed body (unrecognized source), got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

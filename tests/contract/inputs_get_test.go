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
// getInput operation in api/openapi.json (constitution Principle II).

func TestGetInput_RequestAndDecodeContract_Static(t *testing.T) {
	var gotPath string
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1",
			"enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	input, err := hub.GetInput(context.Background(), client, srv.URL, "spotify-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/inputs/spotify-1" {
		t.Errorf("got path %q, want /api/v2/inputs/spotify-1", gotPath)
	}
	if input.InputID != "spotify-1" || input.Source != "STATIC" || input.CreatedAt != nil {
		t.Errorf("unexpected decoded input: %+v", input)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-013)", got)
	}
}

func TestGetInput_RequestAndDecodeContract_Ephemeral(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "line-in-1", "displayName": "Line In", "uri": "u2",
			"enabled": true, "autoRemove": true, "source": "EPHEMERAL", "createdAt": "2026-06-22T14:30:00Z", "pauseable": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	input, err := hub.GetInput(context.Background(), client, srv.URL, "line-in-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Source != "EPHEMERAL" || input.CreatedAt == nil || *input.CreatedAt != "2026-06-22T14:30:00Z" {
		t.Errorf("unexpected decoded input: %+v", input)
	}
}

func TestGetInput_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetInput(context.Background(), client, srv.URL, "spotify-1")
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestGetInput_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetInput(context.Background(), client, srv.URL, "missing-input")
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "input" || notFoundErr.ID != "missing-input" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestGetInput_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"displayName":"Spotify Stream","uri":"u1","enabled":true,"autoRemove":false,"source":"STATIC","createdAt":null,"pauseable":true}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetInput(context.Background(), client, srv.URL, "spotify-1")
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

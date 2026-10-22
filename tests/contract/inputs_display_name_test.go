package contract

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Response/request shapes here mirror #/components/schemas/DisplayNameRequest,
// #/components/schemas/InputResponse, and the setInputDisplayName operation
// in api/openapi.json (constitution Principle II).

func TestSetInputDisplayName_Rename_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "spotify-1", "displayName": "Kiss FM", "uri": "u1",
			"enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true,
		})
	}))
	defer srv.Close()

	name := "Kiss FM"
	input, err := hub.SetInputDisplayName(context.Background(), hub.NewClient(), srv.URL, "spotify-1", &name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", gotMethod)
	}
	if gotPath != "/api/v2/inputs/spotify-1/display-name" {
		t.Errorf("got path %q, want /api/v2/inputs/spotify-1/display-name", gotPath)
	}
	if gotBody != `{"displayName":"Kiss FM"}` {
		t.Errorf("got request body %s, want {\"displayName\":\"Kiss FM\"}", gotBody)
	}
	if input.InputID != "spotify-1" || input.DisplayName != "Kiss FM" {
		t.Errorf("unexpected decoded input: %+v", input)
	}
}

func TestSetInputDisplayName_Reset_SendsNull(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1",
			"enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true,
		})
	}))
	defer srv.Close()

	input, err := hub.SetInputDisplayName(context.Background(), hub.NewClient(), srv.URL, "spotify-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody != `{"displayName":null}` {
		t.Errorf("got request body %s, want {\"displayName\":null}", gotBody)
	}
	if input.DisplayName != "Spotify Stream" {
		t.Errorf("unexpected decoded input: %+v", input)
	}
}

func TestSetInputDisplayName_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "Input not found: missing-input"})
	}))
	defer srv.Close()

	name := "Kiss FM"
	_, err := hub.SetInputDisplayName(context.Background(), hub.NewClient(), srv.URL, "missing-input", &name)
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "input" || notFoundErr.ID != "missing-input" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestSetInputDisplayName_ValidationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Bad Request", "detail": "displayName must be at most 64 characters"})
	}))
	defer srv.Close()

	name := "far too long"
	_, err := hub.SetInputDisplayName(context.Background(), hub.NewClient(), srv.URL, "spotify-1", &name)
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Detail != "displayName must be at most 64 characters" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestSetInputDisplayName_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := hub.SetInputDisplayName(context.Background(), hub.NewClient(), srv.URL, "spotify-1", nil)
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestSetInputDisplayName_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"displayName":"Kiss FM","uri":"u1","enabled":true,"autoRemove":false,"source":"STATIC","createdAt":null,"pauseable":true}`))
	}))
	defer srv.Close()

	_, err := hub.SetInputDisplayName(context.Background(), hub.NewClient(), srv.URL, "spotify-1", nil)
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

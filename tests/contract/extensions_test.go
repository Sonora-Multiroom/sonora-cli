package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"sonora-cli/internal/hub"
)

// Request/response shapes here mirror #/components/schemas/ExtensionInventory
// and #/components/schemas/Extension, and the listExtensions operation, in
// api/openapi.json (constitution Principle II). ListExtensions is used
// internally only, by the TTS "not available" diagnosis (research.md §5).

func TestListExtensions_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/extensions" {
			t.Errorf("got path %q, want /api/v2/extensions", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("got method %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		reason := "bad jar"
		_ = json.NewEncoder(w).Encode(map[string]any{
			"loadingEnabled": true,
			"extensions": []map[string]any{
				{"id": "tts", "name": "Text to Speech", "status": "ACTIVE", "rejectionReason": nil},
				{"id": "other", "name": "Other", "status": "REJECTED", "rejectionReason": reason},
			},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	inv, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.LoadingEnabled == nil || !*inv.LoadingEnabled {
		t.Errorf("expected LoadingEnabled true, got %+v", inv.LoadingEnabled)
	}
	if len(inv.Extensions) != 2 {
		t.Fatalf("expected 2 extensions, got %d: %+v", len(inv.Extensions), inv.Extensions)
	}
	if inv.Extensions[0].ID != "tts" || inv.Extensions[0].Status != "ACTIVE" {
		t.Errorf("unexpected first extension: %+v", inv.Extensions[0])
	}
	if inv.Extensions[0].RejectionReason != nil {
		t.Errorf("expected nil rejection reason for tts, got %v", *inv.Extensions[0].RejectionReason)
	}
	if inv.Extensions[1].RejectionReason == nil || *inv.Extensions[1].RejectionReason != "bad jar" {
		t.Errorf("expected rejection reason %q, got %+v", "bad jar", inv.Extensions[1].RejectionReason)
	}
}

func TestListExtensions_MissingExtensionsArrayDecodesAsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"loadingEnabled": true})
	}))
	defer srv.Close()

	client := hub.NewClient()
	inv, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Extensions) != 0 {
		t.Errorf("expected an empty Extensions slice, got %+v", inv.Extensions)
	}
}

func TestListExtensions_404_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusNotFound {
		t.Errorf("got StatusCode %d, want 404", statusErr.StatusCode)
	}
}

func TestListExtensions_NonJSON200_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a non-JSON 200 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

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

// Response shapes here mirror #/components/schemas/OutputResponse and the
// getOutput operation in api/openapi.json (constitution Principle II).

func TestGetOutput_RequestAndDecodeContract(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"outputId": "office-speaker", "displayName": "Office Speaker",
			"volume": 75, "muted": false, "available": true, "enabled": true,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	output, err := hub.GetOutput(context.Background(), client, srv.URL, "office-speaker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/outputs/office-speaker" {
		t.Errorf("got path %q, want /api/v2/outputs/office-speaker", gotPath)
	}
	if output.OutputID != "office-speaker" || output.DisplayName != "Office Speaker" ||
		output.Volume != 75 || output.Muted || !output.Available || !output.Enabled {
		t.Errorf("unexpected decoded output: %+v", output)
	}
}

func TestGetOutput_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetOutput(context.Background(), client, srv.URL, "office-speaker")
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestGetOutput_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetOutput(context.Background(), client, srv.URL, "missing-speaker")
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
}

func TestGetOutput_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// volume has the wrong JSON type (string instead of int) per FR-013.
		_, _ = w.Write([]byte(`{"outputId":"office-speaker","displayName":"Office Speaker","volume":"loud","muted":false,"available":true,"enabled":true}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetOutput(context.Background(), client, srv.URL, "office-speaker")
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

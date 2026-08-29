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

// Request/response shapes here mirror the deleteInput operation in
// api/openapi.json (constitution Principle II): DELETE /api/v2/inputs/{inputId}
// returns 204 on success, 404 if the input doesn't exist, 400 if it's a
// static (YAML-configured) input.

func TestDeleteInput_Success_204(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteInput(context.Background(), client, srv.URL, "spotify-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("got method %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/v2/inputs/spotify-1" {
		t.Errorf("got path %q, want /api/v2/inputs/spotify-1", gotPath)
	}
}

func TestDeleteInput_404_NotFoundNamesInput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "input not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteInput(context.Background(), client, srv.URL, "missing-input")
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

func TestDeleteInput_400_StaticInputDecodesAsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "urn:multiroom:error:validation-error", "title": "Validation Error", "detail": "static inputs cannot be deleted",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteInput(context.Background(), client, srv.URL, "spotify-1")
	if err == nil {
		t.Fatal("expected an error for a 400 response, got nil")
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Title != "Validation Error" || apiErr.Detail != "static inputs cannot be deleted" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestDeleteInput_400_NonJSONBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteInput(context.Background(), client, srv.URL, "spotify-1")
	if err == nil {
		t.Fatal("expected an error for a non-JSON error body, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError fallback, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusBadRequest {
		t.Errorf("got StatusCode %d, want 400", statusErr.StatusCode)
	}
}

func TestDeleteInput_OtherErrorStatus_IsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteInput(context.Background(), client, srv.URL, "spotify-1")
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

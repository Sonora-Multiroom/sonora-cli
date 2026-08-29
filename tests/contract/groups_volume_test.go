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

// Response/request shapes here mirror #/components/schemas/VolumeRequest,
// #/components/schemas/GroupVolumeResponse, and the setGroupVolume
// operation in api/openapi.json (constitution Principle II).

func TestSetGroupVolume_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groupId": "downstairs", "volume": 75, "updatedAt": "2026-06-22T14:30:00Z",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	gv, err := hub.SetGroupVolume(context.Background(), client, srv.URL, "downstairs", 75)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", gotMethod)
	}
	if gotPath != "/api/v2/groups/downstairs/volume" {
		t.Errorf("got path %q, want /api/v2/groups/downstairs/volume", gotPath)
	}
	if gotBody["volume"] != float64(75) {
		t.Errorf("got request body %+v, want volume=75", gotBody)
	}
	if gv.GroupID != "downstairs" || gv.Volume != 75 || gv.UpdatedAt != "2026-06-22T14:30:00Z" {
		t.Errorf("unexpected decoded group volume: %+v", gv)
	}
}

func TestSetGroupVolume_BoundaryValues(t *testing.T) {
	for _, v := range []int{0, 100} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"groupId": "downstairs", "volume": body["volume"], "updatedAt": "2026-06-22T14:30:00Z",
			})
		}))

		client := hub.NewClient()
		gv, err := hub.SetGroupVolume(context.Background(), client, srv.URL, "downstairs", v)
		srv.Close()
		if err != nil {
			t.Fatalf("volume=%d: unexpected error: %v", v, err)
		}
		if gv.Volume != v {
			t.Errorf("volume=%d: got decoded volume %d", v, gv.Volume)
		}
	}
}

func TestSetGroupVolume_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetGroupVolume(context.Background(), client, srv.URL, "missing-group", 50)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "group" || notFoundErr.ID != "missing-group" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestSetGroupVolume_ValidationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Validation error", "detail": "volume must be between 0 and 100"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetGroupVolume(context.Background(), client, srv.URL, "downstairs", 50)
	if err == nil {
		t.Fatal("expected an error for a 400 response, got nil")
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Detail != "volume must be between 0 and 100" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestSetGroupVolume_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetGroupVolume(context.Background(), client, srv.URL, "downstairs", 50)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestSetGroupVolume_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"volume":75,"updatedAt":"2026-06-22T14:30:00Z"}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetGroupVolume(context.Background(), client, srv.URL, "downstairs", 75)
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

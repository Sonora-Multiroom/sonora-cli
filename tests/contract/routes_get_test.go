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

// Response shapes here mirror #/components/schemas/RouteResponse and the
// getRoute operation in api/openapi.json (constitution Principle II).

func TestGetRoute_RequestAndDecodeContract_PlaybackStarted(t *testing.T) {
	var gotPath string
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"routeId": "route-abc-123", "inputId": "spotify-1", "targetId": "kitchen-speaker",
			"targetType": "SINGLE_OUTPUT", "status": "ACTIVE",
			"createdAt": "2026-06-22T14:30:00Z", "startedAt": "2026-06-22T14:30:01Z",
			"transferable": true, "pauseable": true, "paused": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	route, err := hub.GetRoute(context.Background(), client, srv.URL, "route-abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/routes/route-abc-123" {
		t.Errorf("got path %q, want /api/v2/routes/route-abc-123", gotPath)
	}
	if route.RouteID != "route-abc-123" || route.StartedAt == nil || *route.StartedAt != "2026-06-22T14:30:01Z" {
		t.Errorf("unexpected decoded route: %+v", route)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-012)", got)
	}
}

func TestGetRoute_RequestAndDecodeContract_PlaybackNotStarted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"routeId": "route-def-456", "inputId": "spotify-1", "targetId": "whole-house",
			"targetType": "OUTPUT_GROUP", "status": "STARTING",
			"createdAt": "2026-06-22T14:35:00Z", "startedAt": nil,
			"transferable": false, "pauseable": true, "paused": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	route, err := hub.GetRoute(context.Background(), client, srv.URL, "route-def-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route.StartedAt != nil {
		t.Errorf("expected nil StartedAt, got: %+v", route.StartedAt)
	}
}

func TestGetRoute_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetRoute(context.Background(), client, srv.URL, "route-abc-123")
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestGetRoute_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetRoute(context.Background(), client, srv.URL, "missing-route")
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "route" || notFoundErr.ID != "missing-route" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestGetRoute_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"inputId":"spotify-1","targetId":"kitchen-speaker","targetType":"SINGLE_OUTPUT","status":"ACTIVE","createdAt":"2026-06-22T14:30:00Z","startedAt":null,"transferable":true,"pauseable":true,"paused":false}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetRoute(context.Background(), client, srv.URL, "route-abc-123")
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

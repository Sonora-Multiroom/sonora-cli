package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"sonora-cli/internal/hub"
)

// Response shapes here mirror #/components/schemas/RouteResponse and the
// listRoutes operation in api/openapi.json (constitution Principle II).

func TestListRoutes_RequestAndDecodeContract(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"routeId": "route-abc-123", "inputId": "spotify-1", "targetId": "kitchen-speaker", "targetType": "SINGLE_OUTPUT", "status": "ACTIVE", "createdAt": "2026-06-22T14:30:00Z", "startedAt": "2026-06-22T14:30:01Z", "transferable": true, "pauseable": true, "paused": false},
			{"routeId": "route-def-456", "inputId": "spotify-1", "targetId": "whole-house", "targetType": "OUTPUT_GROUP", "status": "STARTING", "createdAt": "2026-06-22T14:35:00Z", "startedAt": nil, "transferable": false, "pauseable": true, "paused": false},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	routes, err := hub.ListRoutes(context.Background(), client, srv.URL, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/routes" {
		t.Errorf("got path %q, want /api/v2/routes", gotPath)
	}
	if len(gotQuery) != 0 {
		t.Errorf("expected no query parameters when no filters supplied, got: %v", gotQuery)
	}
	if len(routes) != 2 {
		t.Fatalf("got %d routes, want 2", len(routes))
	}
	if routes[0].RouteID != "route-abc-123" || routes[0].StartedAt == nil || *routes[0].StartedAt != "2026-06-22T14:30:01Z" {
		t.Errorf("unexpected decoded route with startedAt populated: %+v", routes[0])
	}
	if routes[1].RouteID != "route-def-456" || routes[1].StartedAt != nil {
		t.Errorf("unexpected decoded route with startedAt null: %+v", routes[1])
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-012)", got)
	}
}

func TestListRoutes_FiltersSetOnlyTheSuppliedQueryParams(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer srv.Close()

	client := hub.NewClient()
	if _, err := hub.ListRoutes(context.Background(), client, srv.URL, "FAILED", "spotify-1", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery.Get("status") != "FAILED" {
		t.Errorf("got status=%q, want %q", gotQuery.Get("status"), "FAILED")
	}
	if gotQuery.Get("inputId") != "spotify-1" {
		t.Errorf("got inputId=%q, want %q", gotQuery.Get("inputId"), "spotify-1")
	}
	if gotQuery.Has("targetId") {
		t.Errorf("expected targetId to be omitted when unset, got: %v", gotQuery)
	}
}

func TestListRoutes_InvalidStatusFilterYieldsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid status"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListRoutes(context.Background(), client, srv.URL, "NOT_A_REAL_STATUS", "", "")
	if err == nil {
		t.Fatal("expected an error for a 400 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestListRoutes_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListRoutes(context.Background(), client, srv.URL, "", "", "")
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestListRoutes_MalformedBodyRejected_MissingRouteID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"inputId":"spotify-1","targetId":"kitchen-speaker","targetType":"SINGLE_OUTPUT","status":"ACTIVE","createdAt":"2026-06-22T14:30:00Z","startedAt":null,"transferable":true,"pauseable":true,"paused":false}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListRoutes(context.Background(), client, srv.URL, "", "", "")
	if err == nil {
		t.Fatal("expected an error for a malformed body (missing routeId), got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestListRoutes_MalformedBodyRejected_UnrecognizedTargetType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"routeId":"route-abc-123","inputId":"spotify-1","targetId":"kitchen-speaker","targetType":"BOGUS","status":"ACTIVE","createdAt":"2026-06-22T14:30:00Z","startedAt":null,"transferable":true,"pauseable":true,"paused":false}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListRoutes(context.Background(), client, srv.URL, "", "", "")
	if err == nil {
		t.Fatal("expected an error for a malformed body (unrecognized targetType), got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestListRoutes_MalformedBodyRejected_UnrecognizedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"routeId":"route-abc-123","inputId":"spotify-1","targetId":"kitchen-speaker","targetType":"SINGLE_OUTPUT","status":"BOGUS","createdAt":"2026-06-22T14:30:00Z","startedAt":null,"transferable":true,"pauseable":true,"paused":false}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListRoutes(context.Background(), client, srv.URL, "", "", "")
	if err == nil {
		t.Fatal("expected an error for a malformed body (unrecognized status), got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

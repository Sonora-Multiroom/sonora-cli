package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Request/response shapes here mirror #/components/schemas/VolumeRequest,
// #/components/schemas/RouteResponse, and the setRouteVolume operation in
// api/openapi.json (constitution Principle II).

func TestSetRouteVolume_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"routeId": "route-3f9a", "inputId": "radio", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT",
			"status": "ACTIVE", "createdAt": "2026-10-25T08:00:00Z", "startedAt": "2026-10-25T08:00:01Z",
			"transferable": true, "pauseable": true, "paused": false, "joinMode": "MIX",
			"outputs": []string{"kitchen"}, "routeVolume": 40, "duckingLevelDb": nil,
		})
	}))
	defer srv.Close()

	route, err := hub.SetRouteVolume(context.Background(), hub.NewClient(), srv.URL, "route-3f9a", 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut || gotPath != "/api/v2/routes/route-3f9a/volume" {
		t.Errorf("got %s %q, want PUT /api/v2/routes/route-3f9a/volume", gotMethod, gotPath)
	}
	if len(gotBody) != 1 || gotBody["volume"] != float64(40) {
		t.Errorf("got request body %+v, want exactly volume=40", gotBody)
	}
	if route.RouteID != "route-3f9a" || route.RouteVolume == nil || *route.RouteVolume != 40 || route.DuckingLevelDb != nil {
		t.Errorf("unexpected decoded route: %+v", route)
	}
}

func TestSetRouteVolume_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "Route not found: missing-route"})
	}))
	defer srv.Close()

	_, err := hub.SetRouteVolume(context.Background(), hub.NewClient(), srv.URL, "missing-route", 40)
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "route" || notFoundErr.ID != "missing-route" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestSetRouteVolume_ValidationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Validation Error", "detail": "volume must be between 0 and 100"})
	}))
	defer srv.Close()

	_, err := hub.SetRouteVolume(context.Background(), hub.NewClient(), srv.URL, "route-3f9a", 40)
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Detail != "volume must be between 0 and 100" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestSetRouteVolume_ValidationErrorWithoutBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	_, err := hub.SetRouteVolume(context.Background(), hub.NewClient(), srv.URL, "route-3f9a", 40)
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected a 400 *hub.StatusError, got %T: %v", err, err)
	}
}

func TestSetRouteVolume_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := hub.SetRouteVolume(context.Background(), hub.NewClient(), srv.URL, "route-3f9a", 40)
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestSetRouteVolume_MalformedBodyRejected(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"not json", `not json`},
		{"missing routeId", `{"inputId": "radio", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT", "status": "ACTIVE", "routeVolume": 40}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			_, err := hub.SetRouteVolume(context.Background(), hub.NewClient(), srv.URL, "route-3f9a", 40)
			var decodeErr *hub.DecodeError
			if !errors.As(err, &decodeErr) {
				t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
			}
		})
	}
}

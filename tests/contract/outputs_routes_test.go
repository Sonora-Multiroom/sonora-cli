package contract

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Response shapes here mirror #/components/schemas/RouteOnOutputResponse and
// the getRoutesForOutput operation in api/openapi.json (constitution
// Principle II).

// routesForOutputServer answers every request with status and body.
func routesForOutputServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGetRoutesForOutput_RequestAndDecodeContract(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"routeId": "route-1", "inputId": "radio", "targetId": "downstairs", "targetType": "OUTPUT_GROUP",
			 "status": "ACTIVE", "joinMode": "REPLACE", "createdAt": "2026-09-30T08:00:00Z",
			 "startedAt": "2026-09-30T08:00:01Z", "transferable": true, "pauseable": true, "paused": false,
			 "duckState": "LOWERED", "main": true},
			{"routeId": "route-2", "inputId": "doorbell", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT",
			 "status": "STARTING", "joinMode": "DUCK_OTHERS", "createdAt": "2026-09-30T08:05:00Z",
			 "startedAt": null, "transferable": false, "pauseable": false, "paused": false,
			 "duckState": "FULL", "main": false}
		]`))
	}))
	defer srv.Close()

	routes, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet || gotPath != "/api/v2/outputs/kitchen/routes" {
		t.Errorf("got %s %q, want GET /api/v2/outputs/kitchen/routes", gotMethod, gotPath)
	}
	if len(routes) != 2 {
		t.Fatalf("got %d routes, want 2: %+v", len(routes), routes)
	}
	startedAt := "2026-09-30T08:00:01Z"
	want0 := hub.RouteOnOutput{
		RouteID: "route-1", InputID: "radio", TargetID: "downstairs", TargetType: "OUTPUT_GROUP",
		Status: "ACTIVE", JoinMode: "REPLACE", CreatedAt: "2026-09-30T08:00:00Z", StartedAt: &startedAt,
		Transferable: true, Pauseable: true, Paused: false, DuckState: "LOWERED", Main: true,
	}
	got0 := routes[0]
	if got0.StartedAt == nil || *got0.StartedAt != startedAt {
		t.Errorf("route 0 startedAt = %v, want %q", got0.StartedAt, startedAt)
	}
	got0.StartedAt, want0.StartedAt = nil, nil
	if got0 != want0 {
		t.Errorf("route 0 = %+v, want %+v", got0, want0)
	}
	want1 := hub.RouteOnOutput{
		RouteID: "route-2", InputID: "doorbell", TargetID: "kitchen", TargetType: "SINGLE_OUTPUT",
		Status: "STARTING", JoinMode: "DUCK_OTHERS", CreatedAt: "2026-09-30T08:05:00Z",
		DuckState: "FULL", Main: false,
	}
	if routes[1] != want1 {
		t.Errorf("route 1 = %+v, want %+v", routes[1], want1)
	}
}

func TestGetRoutesForOutput_EmptyList(t *testing.T) {
	srv := routesForOutputServer(t, http.StatusOK, `[]`)

	routes, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if routes == nil || len(routes) != 0 {
		t.Errorf("got %#v, want a non-nil empty slice", routes)
	}
}

func TestGetRoutesForOutput_NullBody(t *testing.T) {
	srv := routesForOutputServer(t, http.StatusOK, `null`)

	routes, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if routes == nil || len(routes) != 0 {
		t.Errorf("got %#v, want a non-nil empty slice", routes)
	}
}

func TestGetRoutesForOutput_NotFound(t *testing.T) {
	srv := routesForOutputServer(t, http.StatusNotFound, `{"title": "Not Found", "detail": "Output not found: kitchen"}`)

	_, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "output" || notFoundErr.ID != "kitchen" {
		t.Errorf("got %+v, want Resource=output ID=kitchen", notFoundErr)
	}
}

func TestGetRoutesForOutput_MethodNotAllowed(t *testing.T) {
	srv := routesForOutputServer(t, http.StatusMethodNotAllowed, ``)

	_, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("got status %d, want 405", statusErr.StatusCode)
	}
}

const routeOnOutputTemplate = `[{"routeId": "route-1", "inputId": "radio", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT",
	"status": "ACTIVE", "joinMode": %q, "createdAt": "2026-09-30T08:00:00Z", "startedAt": null,
	"transferable": true, "pauseable": true, "paused": false, "duckState": %q, "main": true}]`

func TestGetRoutesForOutput_UnknownDuckState(t *testing.T) {
	srv := routesForOutputServer(t, http.StatusOK, fmt.Sprintf(routeOnOutputTemplate, "REPLACE", "MUFFLED"))

	routes, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if routes[0].DuckState != "MUFFLED" {
		t.Errorf("duckState = %q, want MUFFLED unchanged", routes[0].DuckState)
	}
	if routes[0].Lowered() {
		t.Error("expected an unknown duck state not to count as lowered")
	}
}

func TestGetRoutesForOutput_UnknownJoinMode(t *testing.T) {
	srv := routesForOutputServer(t, http.StatusOK, fmt.Sprintf(routeOnOutputTemplate, "CROSSFADE", "FULL"))

	routes, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if routes[0].JoinMode != "CROSSFADE" {
		t.Errorf("joinMode = %q, want CROSSFADE unchanged", routes[0].JoinMode)
	}
}

func TestGetRoutesForOutput_MalformedBodyRejected(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"not an array", `{"not": "an array"}`},
		{"missing routeId", `[{"inputId": "radio", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT", "status": "ACTIVE"}]`},
		{"unknown targetType", `[{"routeId": "r", "inputId": "radio", "targetId": "kitchen", "targetType": "SPEAKER", "status": "ACTIVE"}]`},
		{"unknown status", `[{"routeId": "r", "inputId": "radio", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT", "status": "PLAYING"}]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := routesForOutputServer(t, http.StatusOK, tt.body)

			_, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
			var decodeErr *hub.DecodeError
			if !errors.As(err, &decodeErr) {
				t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
			}
		})
	}
}

func TestGetRoutesForOutput_DecodesRouteVolumeAndDuckingLevel(t *testing.T) {
	srv := routesForOutputServer(t, http.StatusOK, `[
		{"routeId": "route-1", "inputId": "radio", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT",
		 "status": "ACTIVE", "joinMode": "REPLACE", "createdAt": "2026-09-30T08:00:00Z", "startedAt": null,
		 "transferable": true, "pauseable": true, "paused": false, "duckState": "LOWERED", "main": true,
		 "routeVolume": null, "duckingLevelDb": null},
		{"routeId": "route-2", "inputId": "doorbell", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT",
		 "status": "ACTIVE", "joinMode": "DUCK_OTHERS", "createdAt": "2026-09-30T08:05:00Z", "startedAt": null,
		 "transferable": false, "pauseable": false, "paused": false, "duckState": "FULL", "main": false,
		 "routeVolume": 100, "duckingLevelDb": -25.0}
	]`)

	routes, err := hub.GetRoutesForOutput(context.Background(), hub.NewClient(), srv.URL, "kitchen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if routes[0].RouteVolume != nil || routes[0].DuckingLevelDb != nil {
		t.Errorf("route 0 RouteVolume/DuckingLevelDb = %v/%v, want nil/nil", routes[0].RouteVolume, routes[0].DuckingLevelDb)
	}
	if routes[1].RouteVolume == nil || *routes[1].RouteVolume != 100 || routes[1].DuckingLevelDb == nil || *routes[1].DuckingLevelDb != -25 {
		t.Errorf("route 1 RouteVolume/DuckingLevelDb = %v/%v, want 100/-25", routes[1].RouteVolume, routes[1].DuckingLevelDb)
	}
}

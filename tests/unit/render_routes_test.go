package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func TestRenderRoutesYAML_RendersListViewFieldsInOrder(t *testing.T) {
	routes := []hub.Route{
		{
			RouteID: "route-abc-123", InputID: "spotify-1", TargetID: "kitchen-speaker",
			TargetType: "SINGLE_OUTPUT", Status: "ACTIVE",
			CreatedAt: "2026-06-22T14:30:00Z", StartedAt: strPtr("2026-06-22T14:30:01Z"),
			Transferable: true, Pauseable: true, Paused: false,
		},
	}
	got := render.RenderRoutesYAML(routes)

	for _, want := range []string{
		`routeId: "route-abc-123"`,
		`inputId: "spotify-1"`,
		`targetId: "kitchen-speaker"`,
		`targetType: "SINGLE_OUTPUT"`,
		`status: "ACTIVE"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("YAML output missing %q; got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"createdAt", "startedAt", "transferable", "pauseable", "paused"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("list view must not include %q; got:\n%s", unwanted, got)
		}
	}

	routeIdx := strings.Index(got, "routeId:")
	inputIdx := strings.Index(got, "inputId:")
	targetIdx := strings.Index(got, "targetId:")
	targetTypeIdx := strings.Index(got, "targetType:")
	statusIdx := strings.Index(got, "status:")
	if !(routeIdx < inputIdx && inputIdx < targetIdx && targetIdx < targetTypeIdx && targetTypeIdx < statusIdx) {
		t.Errorf("expected documented field order, got:\n%s", got)
	}
}

func TestRenderRoutesYAML_ZeroRoutesIsUnambiguous(t *testing.T) {
	got := render.RenderRoutesYAML(nil)

	if !strings.Contains(got, "routes: []") {
		t.Errorf("expected an explicit empty routes list; got:\n%s", got)
	}
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "no routes") {
		t.Errorf("expected an unambiguous 'no routes' note; got:\n%s", got)
	}
}

func TestRenderRouteYAML_AllTenFieldsAsBareRecord(t *testing.T) {
	r := hub.Route{
		RouteID: "route-abc-123", InputID: "spotify-1", TargetID: "kitchen-speaker",
		TargetType: "SINGLE_OUTPUT", Status: "ACTIVE",
		CreatedAt: "2026-06-22T14:30:00Z", StartedAt: strPtr("2026-06-22T14:30:01Z"),
		Transferable: true, Pauseable: true, Paused: false,
	}
	got := render.RenderRouteYAML(r)

	if strings.Contains(got, "routes:") {
		t.Errorf("expected a bare record, not a list wrapper, got:\n%s", got)
	}
	for _, want := range []string{
		`routeId: "route-abc-123"`, `inputId: "spotify-1"`, `targetId: "kitchen-speaker"`,
		`targetType: "SINGLE_OUTPUT"`, `status: "ACTIVE"`,
		`createdAt: "2026-06-22T14:30:00Z"`, `startedAt: "2026-06-22T14:30:01Z"`,
		"transferable: true", "pauseable: true", "paused: false",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderRouteYAML_StartedAtNullBareUnquotedWhenNotStarted(t *testing.T) {
	r := hub.Route{
		RouteID: "route-def-456", InputID: "spotify-1", TargetID: "whole-house",
		TargetType: "OUTPUT_GROUP", Status: "STARTING",
		CreatedAt: "2026-06-22T14:35:00Z", StartedAt: nil,
		Transferable: false, Pauseable: true, Paused: false,
	}
	got := render.RenderRouteYAML(r)

	if !strings.Contains(got, "startedAt: null") {
		t.Errorf("expected bare unquoted 'startedAt: null', got:\n%s", got)
	}
	if strings.Contains(got, `startedAt: "null"`) {
		t.Errorf("startedAt: null must not be quoted, got:\n%s", got)
	}
}

func TestRenderRoutesJSON_StrictlyParseable(t *testing.T) {
	routes := []hub.Route{
		{
			RouteID: "route-abc-123", InputID: "spotify-1", TargetID: "kitchen-speaker",
			TargetType: "SINGLE_OUTPUT", Status: "ACTIVE",
			CreatedAt: "2026-06-22T14:30:00Z", StartedAt: strPtr("2026-06-22T14:30:01Z"),
			Transferable: true, Pauseable: true, Paused: false,
		},
	}
	got := render.RenderRoutesJSON(routes)

	var decoded struct {
		Routes []struct {
			RouteID    string `json:"routeId"`
			InputID    string `json:"inputId"`
			TargetID   string `json:"targetId"`
			TargetType string `json:"targetType"`
			Status     string `json:"status"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if len(decoded.Routes) != 1 || decoded.Routes[0].RouteID != "route-abc-123" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
	if strings.Contains(got, "createdAt") || strings.Contains(got, "startedAt") {
		t.Errorf("list JSON must not include get-only fields, got: %s", got)
	}
}

func TestRenderRoutesJSON_ZeroRoutes(t *testing.T) {
	got := render.RenderRoutesJSON([]hub.Route(nil))

	var decoded struct {
		Routes []hub.Route `json:"routes"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}
	if decoded.Routes == nil || len(decoded.Routes) != 0 {
		t.Errorf("expected an explicit empty array, got: %s", got)
	}
}

func TestRenderRouteJSON_SingleObjectRoundTrips(t *testing.T) {
	r := hub.Route{
		RouteID: "route-abc-123", InputID: "spotify-1", TargetID: "kitchen-speaker",
		TargetType: "SINGLE_OUTPUT", Status: "ACTIVE",
		CreatedAt: "2026-06-22T14:30:00Z", StartedAt: strPtr("2026-06-22T14:30:01Z"),
		Transferable: true, Pauseable: true, Paused: false,
	}
	got := render.RenderRouteJSON(r)

	if strings.Contains(got, `"routes"`) || strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Errorf("expected a single object, not a list wrapper, got:\n%s", got)
	}

	var decoded hub.Route
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output did not round-trip through json.Unmarshal: %v\ngot: %s", err, got)
	}
	if decoded.RouteID != r.RouteID || decoded.CreatedAt != r.CreatedAt || *decoded.StartedAt != *r.StartedAt {
		t.Errorf("round-tripped value = %+v, want %+v", decoded, r)
	}
}

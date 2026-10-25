package unit

import (
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func TestRenderRouteYAML_RouteVolumeAndDuckingLevelAfterJoinMode(t *testing.T) {
	r := joinModeRouteFixture()
	r.JoinMode = "DUCK_OTHERS"
	r.RouteVolume = new(100)
	r.DuckingLevelDb = new(-12.5)
	r.Outputs = []string{"kitchen"}

	got := render.RenderRouteYAML(r)
	want := "joinMode: \"DUCK_OTHERS\"\nrouteVolume: 100\nduckingLevelDb: -12.5\noutputs:\n  - \"kitchen\"\n"
	if !strings.HasSuffix(got, want) {
		t.Errorf("expected the record to end with:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderRouteYAML_WholeDecibelsPrintWithoutDecimals(t *testing.T) {
	r := joinModeRouteFixture()
	r.JoinMode = "DUCK_OTHERS"
	r.DuckingLevelDb = new(-25.0)

	if got := render.RenderRouteYAML(r); !strings.Contains(got, "\nduckingLevelDb: -25\n") {
		t.Errorf("expected duckingLevelDb: -25, got:\n%s", got)
	}
}

func TestRenderRouteYAML_NilRouteVolumeAndDuckingLevelLeftOut(t *testing.T) {
	r := joinModeRouteFixture()
	r.JoinMode = "MIX"

	got := render.RenderRouteYAML(r)
	if strings.Contains(got, "routeVolume") || strings.Contains(got, "duckingLevelDb") {
		t.Errorf("expected no routeVolume or duckingLevelDb line, got:\n%s", got)
	}
}

func TestRenderRoutesYAML_ListUnchangedByRouteVolume(t *testing.T) {
	r := joinModeRouteFixture()
	r.RouteVolume = new(40)

	if got := render.RenderRoutesYAML([]hub.Route{r}); strings.Contains(got, "routeVolume") {
		t.Errorf("expected the list view without routeVolume, got:\n%s", got)
	}
}

func TestRenderOutputWithRoutesYAML_RouteVolumeAndDuckingLevelAfterJoinMode(t *testing.T) {
	o, routes := outputWithRoutesFixture()
	routes[1].RouteVolume = new(90)
	routes[1].DuckingLevelDb = new(-25.0)

	got := render.RenderOutputWithRoutesYAML(o, routes)
	want := `    joinMode: "DUCK_OTHERS"
    routeVolume: 90
    duckingLevelDb: -25
    duckState: "FULL"
`
	if !strings.Contains(got, want) {
		t.Errorf("expected route 2 to contain:\n%s\ngot:\n%s", want, got)
	}
	route1, _, _ := strings.Cut(got, "  - routeId: \"route-2\"")
	if strings.Contains(route1, "routeVolume") || strings.Contains(route1, "duckingLevelDb") {
		t.Errorf("expected route 1 without routeVolume or duckingLevelDb, got:\n%s", got)
	}
}

func TestRenderInputYAML_RouteDefaultsAfterDefaultJoinMode(t *testing.T) {
	mode := "DUCK_OTHERS"
	i := hub.Input{
		InputID: "doorbell", DisplayName: "Doorbell", Source: "STATIC",
		DefaultJoinMode: &mode, DefaultRouteVolume: new(90), DefaultDuckingLevelDb: new(-30.0),
	}

	got := render.RenderInputYAML(i)
	want := "defaultJoinMode: \"DUCK_OTHERS\"\ndefaultRouteVolume: 90\ndefaultDuckingLevelDb: -30\n"
	if !strings.HasSuffix(got, want) {
		t.Errorf("expected the record to end with:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderInputYAML_NilRouteDefaultsLeftOut(t *testing.T) {
	i := hub.Input{InputID: "spotify-1", DisplayName: "Spotify Stream", Source: "STATIC"}

	got := render.RenderInputYAML(i)
	if strings.Contains(got, "defaultRouteVolume") || strings.Contains(got, "defaultDuckingLevelDb") {
		t.Errorf("expected no route default lines, got:\n%s", got)
	}
}

func TestRenderInputsYAML_ListUnchangedByRouteDefaults(t *testing.T) {
	i := hub.Input{InputID: "doorbell", DisplayName: "Doorbell", Source: "STATIC", DefaultRouteVolume: new(90)}

	if got := render.RenderInputsYAML([]hub.Input{i}); strings.Contains(got, "defaultRouteVolume") {
		t.Errorf("expected the list view without defaultRouteVolume, got:\n%s", got)
	}
}

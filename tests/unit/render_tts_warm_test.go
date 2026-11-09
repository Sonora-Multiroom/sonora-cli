package unit

import (
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func TestRenderWarmYAML(t *testing.T) {
	if got, want := render.RenderWarmYAML(hub.WarmResult{CacheHit: false}), "cacheHit: false\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := render.RenderWarmYAML(hub.WarmResult{CacheHit: true}), "cacheHit: true\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderWarmJSON(t *testing.T) {
	if got, want := render.RenderWarmJSON(hub.WarmResult{CacheHit: false}), "{\"cacheHit\":false}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := render.RenderWarmJSON(hub.WarmResult{CacheHit: true}), "{\"cacheHit\":true}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

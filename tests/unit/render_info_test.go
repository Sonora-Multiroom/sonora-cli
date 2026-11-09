package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func i64(v int64) *int64 { return &v }

func TestRenderInfoYAML_AllFields(t *testing.T) {
	got := render.RenderInfoYAML(hub.Info{
		Product: "multiroom", Version: "0.1.33", APIVersion: "0.1.33",
		StartedAt: "2026-11-06T10:06:59.512Z", UptimeSeconds: i64(16981),
	})
	want := "product: \"multiroom\"\nversion: \"0.1.33\"\napiVersion: \"0.1.33\"\n" +
		"startedAt: \"2026-11-06T10:06:59.512Z\"\nuptimeSeconds: 16981\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderInfoYAML_OmitsAbsentFields(t *testing.T) {
	got := render.RenderInfoYAML(hub.Info{Product: "multiroom", Version: "unknown", APIVersion: "unknown"})
	if strings.Contains(got, "startedAt") || strings.Contains(got, "uptimeSeconds") {
		t.Errorf("absent fields must be left out, got:\n%s", got)
	}
	if !strings.Contains(got, "version: \"unknown\"\n") {
		t.Errorf("got:\n%s", got)
	}
}

func TestRenderInfoYAML_ZeroUptimeShown(t *testing.T) {
	got := render.RenderInfoYAML(hub.Info{Product: "multiroom", Version: "1", APIVersion: "1", UptimeSeconds: i64(0)})
	if !strings.Contains(got, "uptimeSeconds: 0\n") {
		t.Errorf("a zero uptime the hub sent must be shown, got:\n%s", got)
	}
}

func TestRenderInfoJSON(t *testing.T) {
	got := render.RenderInfoJSON(hub.Info{
		Product: "multiroom", Version: "0.1.33", APIVersion: "0.1.33",
		StartedAt: "2026-11-06T10:06:59.512Z", UptimeSeconds: i64(16981),
	})
	want := `{"product":"multiroom","version":"0.1.33","apiVersion":"0.1.33","startedAt":"2026-11-06T10:06:59.512Z","uptimeSeconds":16981}` + "\n"
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatal(err)
	}
	got = render.RenderInfoJSON(hub.Info{Product: "multiroom", Version: "1", APIVersion: "1"})
	if strings.Contains(got, "startedAt") || strings.Contains(got, "uptimeSeconds") {
		t.Errorf("absent fields must be left out of JSON, got %s", got)
	}
}

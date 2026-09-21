package unit

import (
	"encoding/json"
	"testing"

	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

func TestRenderSpeakYAML_ExactFieldsAndOrder(t *testing.T) {
	got := render.RenderSpeakYAML(hub.SpeakAccepted{
		AnnouncementID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890", CacheHit: true, QueueDepth: 1,
	})
	want := "announcementId: \"a1b2c3d4-e5f6-7890-abcd-ef1234567890\"\ncacheHit: true\nqueueDepth: 1\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRenderSpeakJSON_ExactKeys(t *testing.T) {
	got := render.RenderSpeakJSON(hub.SpeakAccepted{
		AnnouncementID: "a1", CacheHit: false, QueueDepth: 3,
	})
	want := `{"announcementId":"a1","cacheHit":false,"queueDepth":3}` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(got[:len(got)-1]), &decoded); err != nil {
		t.Fatalf("rendered JSON is invalid: %v", err)
	}
	if len(decoded) != 3 {
		t.Errorf("expected exactly 3 keys, got: %+v", decoded)
	}
}

func TestRenderTTSCacheStatsYAML_PopulatedMap_SortedAndQuoted(t *testing.T) {
	got := render.RenderTTSCacheStatsYAML(hub.TTSCacheStats{
		TotalEntries: 42, TotalSizeBytes: 15728640, MaxSizeBytes: 524288000,
		EntriesByProvider: map[string]int32{"piper-local": 7, "openai": 35},
	})
	want := "totalEntries: 42\ntotalSizeBytes: 15728640\nmaxSizeBytes: 524288000\n" +
		"entriesByProvider:\n  \"openai\": 35\n  \"piper-local\": 7\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRenderTTSCacheStatsYAML_EmptyMap(t *testing.T) {
	got := render.RenderTTSCacheStatsYAML(hub.TTSCacheStats{
		TotalEntries: 0, TotalSizeBytes: 0, MaxSizeBytes: 524288000,
		EntriesByProvider: map[string]int32{},
	})
	want := "totalEntries: 0\ntotalSizeBytes: 0\nmaxSizeBytes: 524288000\nentriesByProvider: {}\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRenderTTSCacheStatsJSON_AlwaysHasEntriesByProviderObject(t *testing.T) {
	gotEmpty := render.RenderTTSCacheStatsJSON(hub.TTSCacheStats{EntriesByProvider: nil})
	var decodedEmpty map[string]any
	if err := json.Unmarshal([]byte(gotEmpty[:len(gotEmpty)-1]), &decodedEmpty); err != nil {
		t.Fatalf("rendered JSON is invalid: %v", err)
	}
	m, ok := decodedEmpty["entriesByProvider"].(map[string]any)
	if !ok || len(m) != 0 {
		t.Errorf("expected entriesByProvider: {}, got: %+v", decodedEmpty["entriesByProvider"])
	}

	gotPopulated := render.RenderTTSCacheStatsJSON(hub.TTSCacheStats{
		EntriesByProvider: map[string]int32{"openai": 35},
	})
	var decodedPopulated map[string]any
	if err := json.Unmarshal([]byte(gotPopulated[:len(gotPopulated)-1]), &decodedPopulated); err != nil {
		t.Fatalf("rendered JSON is invalid: %v", err)
	}
	mp, ok := decodedPopulated["entriesByProvider"].(map[string]any)
	if !ok || mp["openai"] != float64(35) {
		t.Errorf("expected entriesByProvider.openai == 35, got: %+v", decodedPopulated["entriesByProvider"])
	}
}

func TestRenderTTSCacheClearedYAML_All(t *testing.T) {
	got := render.RenderTTSCacheClearedYAML(nil)
	want := "cleared: all\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderTTSCacheClearedYAML_Provider(t *testing.T) {
	provider := "openai"
	got := render.RenderTTSCacheClearedYAML(&provider)
	want := "cleared: provider\nprovider: \"openai\"\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderTTSCacheClearedJSON_All(t *testing.T) {
	got := render.RenderTTSCacheClearedJSON(nil)
	want := `{"cleared":"all"}` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderTTSCacheClearedJSON_Provider(t *testing.T) {
	provider := "openai"
	got := render.RenderTTSCacheClearedJSON(&provider)
	want := `{"cleared":"provider","provider":"openai"}` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderTTSCacheStatsYAML_LargeSizeRendersExactly(t *testing.T) {
	got := render.RenderTTSCacheStatsYAML(hub.TTSCacheStats{
		TotalEntries: 1, TotalSizeBytes: 1 << 33, MaxSizeBytes: 1 << 34,
		EntriesByProvider: map[string]int32{},
	})
	want := "totalEntries: 1\ntotalSizeBytes: 8589934592\nmaxSizeBytes: 17179869184\nentriesByProvider: {}\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

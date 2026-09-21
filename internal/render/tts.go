package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"sonora-cli/internal/hub"
)

// speakPayload is the flat rendered view of a hub.SpeakAccepted: exactly
// announcementId, cacheHit, queueDepth, in that order (data-model.md's
// "Rendered views" section).
type speakPayload struct {
	AnnouncementID string `json:"announcementId"`
	CacheHit       bool   `json:"cacheHit"`
	QueueDepth     int32  `json:"queueDepth"`
}

func toSpeakPayload(s hub.SpeakAccepted) speakPayload {
	return speakPayload{AnnouncementID: s.AnnouncementID, CacheHit: s.CacheHit, QueueDepth: s.QueueDepth}
}

// RenderSpeakYAML renders a speak result as a bare YAML record, exposing
// exactly announcementId, cacheHit, queueDepth in that order.
func RenderSpeakYAML(s hub.SpeakAccepted) string {
	payload := toSpeakPayload(s)
	var b bytes.Buffer
	fmt.Fprintf(&b, "announcementId: %q\n", payload.AnnouncementID)
	fmt.Fprintf(&b, "cacheHit: %t\n", payload.CacheHit)
	fmt.Fprintf(&b, "queueDepth: %d\n", payload.QueueDepth)
	return b.String()
}

// RenderSpeakJSON renders a speak result as a strict JSON object, exposing
// exactly announcementId, cacheHit, queueDepth.
func RenderSpeakJSON(s hub.SpeakAccepted) string {
	data, err := json.Marshal(toSpeakPayload(s))
	if err != nil {
		// speakPayload's fields are plain string/bool/int32 — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// cacheStatsPayload is the flat rendered view of a hub.TTSCacheStats:
// exactly totalEntries, totalSizeBytes, maxSizeBytes, entriesByProvider, in
// that order (data-model.md's "Rendered views" section). EntriesByProvider
// is never nil, so it always marshals as a JSON object.
type cacheStatsPayload struct {
	TotalEntries      int32            `json:"totalEntries"`
	TotalSizeBytes    int64            `json:"totalSizeBytes"`
	MaxSizeBytes      int64            `json:"maxSizeBytes"`
	EntriesByProvider map[string]int32 `json:"entriesByProvider"`
}

func toCacheStatsPayload(s hub.TTSCacheStats) cacheStatsPayload {
	entries := s.EntriesByProvider
	if entries == nil {
		entries = map[string]int32{}
	}
	return cacheStatsPayload{
		TotalEntries: s.TotalEntries, TotalSizeBytes: s.TotalSizeBytes, MaxSizeBytes: s.MaxSizeBytes,
		EntriesByProvider: entries,
	}
}

// RenderTTSCacheStatsYAML renders cache statistics as a bare YAML record.
// entriesByProvider is a nested mapping with provider names quoted and
// sorted for stable output, or the literal `{}` when empty.
func RenderTTSCacheStatsYAML(s hub.TTSCacheStats) string {
	payload := toCacheStatsPayload(s)
	var b bytes.Buffer
	fmt.Fprintf(&b, "totalEntries: %d\n", payload.TotalEntries)
	fmt.Fprintf(&b, "totalSizeBytes: %d\n", payload.TotalSizeBytes)
	fmt.Fprintf(&b, "maxSizeBytes: %d\n", payload.MaxSizeBytes)
	if len(payload.EntriesByProvider) == 0 {
		fmt.Fprint(&b, "entriesByProvider: {}\n")
		return b.String()
	}
	fmt.Fprint(&b, "entriesByProvider:\n")
	providers := make([]string, 0, len(payload.EntriesByProvider))
	for name := range payload.EntriesByProvider {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	for _, name := range providers {
		fmt.Fprintf(&b, "  %q: %d\n", name, payload.EntriesByProvider[name])
	}
	return b.String()
}

// RenderTTSCacheStatsJSON renders cache statistics as a strict JSON object.
// entriesByProvider is always present as an object, `{}` when empty, never
// omitted or null.
func RenderTTSCacheStatsJSON(s hub.TTSCacheStats) string {
	data, err := json.Marshal(toCacheStatsPayload(s))
	if err != nil {
		// cacheStatsPayload's fields are plain int32/int64/map[string]int32 —
		// Marshal cannot fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

// ttsCacheClearedPayload is the flat rendered view of a `clear tts-cache`
// result: `cleared` plus `provider` (omitted when clearing all providers).
type ttsCacheClearedPayload struct {
	Cleared  string  `json:"cleared"`
	Provider *string `json:"provider,omitempty"`
}

func toTTSCacheClearedPayload(provider *string) ttsCacheClearedPayload {
	if provider == nil {
		return ttsCacheClearedPayload{Cleared: "all"}
	}
	return ttsCacheClearedPayload{Cleared: "provider", Provider: provider}
}

// RenderTTSCacheClearedYAML renders a `clear tts-cache` result: `cleared:
// all`, or `cleared: provider` followed by `provider: "<name>"`.
func RenderTTSCacheClearedYAML(provider *string) string {
	payload := toTTSCacheClearedPayload(provider)
	var b bytes.Buffer
	fmt.Fprintf(&b, "cleared: %s\n", payload.Cleared)
	if payload.Provider != nil {
		fmt.Fprintf(&b, "provider: %q\n", *payload.Provider)
	}
	return b.String()
}

// RenderTTSCacheClearedJSON renders a `clear tts-cache` result as a strict
// JSON object: `{"cleared":"all"}` or `{"cleared":"provider","provider":"<name>"}`.
func RenderTTSCacheClearedJSON(provider *string) string {
	data, err := json.Marshal(toTTSCacheClearedPayload(provider))
	if err != nil {
		// ttsCacheClearedPayload's fields are plain string/*string — Marshal
		// cannot fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

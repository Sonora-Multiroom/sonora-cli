package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sonora-cli/internal/hub"
)

// playbackPayload is the flat rendered view of a hub.PlaybackResponse: only
// inputId/routeId/status/message (FR-006) — the full nested Route is
// deliberately not rendered (data-model.md's "PlaybackResult" section).
type playbackPayload struct {
	InputID string `json:"inputId"`
	RouteID string `json:"routeId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func toPlaybackPayload(p hub.PlaybackResponse) playbackPayload {
	return playbackPayload{
		InputID: p.InputID,
		RouteID: p.Route.RouteID,
		Status:  p.Route.Status,
		Message: p.Message,
	}
}

// RenderPlaybackYAML renders a playback result as a bare YAML record,
// exposing exactly inputId, routeId, status, message in that order (FR-006).
func RenderPlaybackYAML(p hub.PlaybackResponse) string {
	payload := toPlaybackPayload(p)
	var b bytes.Buffer
	fmt.Fprintf(&b, "inputId: %q\n", payload.InputID)
	fmt.Fprintf(&b, "routeId: %q\n", payload.RouteID)
	fmt.Fprintf(&b, "status: %q\n", payload.Status)
	fmt.Fprintf(&b, "message: %q\n", payload.Message)
	return b.String()
}

// RenderPlaybackJSON renders a playback result as a strict JSON object,
// exposing exactly inputId, routeId, status, message (FR-006).
func RenderPlaybackJSON(p hub.PlaybackResponse) string {
	data, err := json.Marshal(toPlaybackPayload(p))
	if err != nil {
		// playbackPayload's fields are all plain strings — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

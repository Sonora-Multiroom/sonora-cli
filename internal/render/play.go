package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// playbackPayload is the flat rendered view of a hub.PlaybackResponse:
// inputId, displayName (only when the hub reports one), routeId, status and
// message. The full nested Route is deliberately not rendered
// (data-model.md's "PlaybackResult" section).
type playbackPayload struct {
	InputID     string `json:"inputId"`
	DisplayName string `json:"displayName,omitempty"`
	RouteID     string `json:"routeId"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

func toPlaybackPayload(p hub.PlaybackResponse) playbackPayload {
	return playbackPayload{
		InputID:     p.InputID,
		DisplayName: p.DisplayName,
		RouteID:     p.Route.RouteID,
		Status:      p.Route.Status,
		Message:     p.Message,
	}
}

// RenderPlaybackYAML renders a playback result as a bare YAML record,
// exposing inputId, displayName, routeId, status, message in that order;
// displayName is left out when the hub did not report one.
func RenderPlaybackYAML(p hub.PlaybackResponse) string {
	payload := toPlaybackPayload(p)
	var b bytes.Buffer
	fmt.Fprintf(&b, "inputId: %q\n", payload.InputID)
	if payload.DisplayName != "" {
		fmt.Fprintf(&b, "displayName: %q\n", payload.DisplayName)
	}
	fmt.Fprintf(&b, "routeId: %q\n", payload.RouteID)
	fmt.Fprintf(&b, "status: %q\n", payload.Status)
	fmt.Fprintf(&b, "message: %q\n", payload.Message)
	return b.String()
}

// RenderPlaybackJSON renders a playback result as a strict JSON object,
// exposing inputId, displayName, routeId, status, message; displayName is
// left out when the hub did not report one.
func RenderPlaybackJSON(p hub.PlaybackResponse) string {
	data, err := json.Marshal(toPlaybackPayload(p))
	if err != nil {
		// playbackPayload's fields are all plain strings — Marshal cannot
		// fail for this input shape.
		panic(err)
	}
	return string(data) + "\n"
}

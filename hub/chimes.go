package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ChimeNone is the reserved chime value that means no chime, whatever a
// lower level (an input's default, an output's or a group's setting) says. It
// is matched exactly, in lower case, and is never listed by ListChimes.
const ChimeNone = "none"

// Chime is a short sound the hub has loaded, to play before a route's audio.
// It mirrors #/components/schemas/ChimeResponse. ID is what to send as a
// request's Chime or an input's DefaultChime.
type Chime struct {
	ID             string `json:"id"`
	DurationMillis int    `json:"durationMillis"`
}

// ListChimes calls GET {baseURL}/api/v2/chimes (operationId "listChimes") and
// returns every chime the hub has loaded, in the hub's configuration order.
// The list is empty, not nil, when none is configured. A hub older than 0.1.34
// answers 404, returned as a *StatusError like any other non-2xx status; a
// body that is not a JSON object, or that lists a chime without an id, is a
// *DecodeError.
func ListChimes(ctx context.Context, client *http.Client, baseURL string) ([]Chime, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/api/v2/chimes", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var wire struct {
		Chimes []Chime `json:"chimes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wire); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if wire.Chimes == nil {
		wire.Chimes = []Chime{}
	}
	for _, c := range wire.Chimes {
		if c.ID == "" {
			return nil, &DecodeError{Err: fmt.Errorf("chimes response lists a chime without an id")}
		}
	}
	return wire.Chimes, nil
}

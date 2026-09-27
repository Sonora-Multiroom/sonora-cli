// Package hub is the Multiroom Audio Hub API client.
package hub

import (
	"net/http"
	"time"
)

// requestTimeout bounds the full round trip of a single request (constitution
// Principle IV) — no unbounded waits, no automatic retries.
const requestTimeout = 5 * time.Second

// NewClient returns an HTTP client bound to the hub with a fixed overall
// request timeout and the default transport (connection reuse). It performs
// no I/O itself, so it is safe to construct at the point a command handler
// needs it, after argument parsing has completed (constitution Principle I).
func NewClient() *http.Client {
	return NewClientWithTimeout(requestTimeout)
}

// NewClientWithTimeout returns an HTTP client bound to the hub like
// NewClient, but with an overall request timeout of d instead of the
// standard 5s. `speak` uses this with hub.SpeakTimeout (15s) so the hub's
// own provider-timeout error arrives before the CLI's request bound does
// (research.md §6); every other command keeps the 5s default via NewClient.
func NewClientWithTimeout(d time.Duration) *http.Client {
	return &http.Client{
		Timeout:   d,
		Transport: http.DefaultTransport,
	}
}

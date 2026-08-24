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
	return &http.Client{
		Timeout:   requestTimeout,
		Transport: http.DefaultTransport,
	}
}

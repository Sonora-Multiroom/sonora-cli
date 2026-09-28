package hub

import (
	"net/http"
	"time"
)

// requestTimeout bounds the full round trip of a single request — no
// unbounded waits, no automatic retries.
const requestTimeout = 5 * time.Second

// NewClient returns an HTTP client bound to the hub with a fixed overall
// request timeout and the default transport (connection reuse). It performs
// no I/O itself, so it is safe to construct right before use.
func NewClient() *http.Client {
	return NewClientWithTimeout(requestTimeout)
}

// NewClientWithTimeout returns an HTTP client bound to the hub like
// NewClient, but with an overall request timeout of d instead of the
// standard 5s. Use a longer timeout for a call expected to take longer than
// 5s, such as Speak (see SpeakTimeout), so the hub's own timeout error
// arrives before this client's request bound does.
func NewClientWithTimeout(d time.Duration) *http.Client {
	return &http.Client{
		Timeout:   d,
		Transport: http.DefaultTransport,
	}
}

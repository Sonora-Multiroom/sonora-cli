// Package hub is a client for the Multiroom Audio Hub API described by
// api/openapi.json.
//
// Construct an *http.Client with NewClient (or NewClientWithTimeout for a
// non-default deadline), then call one of the operation functions —
// ListOutputs, GetGroup, Playback, Speak, and so on — passing a context, the
// client, and the hub's base URL. Each call makes exactly one HTTP request:
// there are no retries and no hidden background work.
//
// Every operation returns a typed error (*StatusError, *APIError,
// *NotFoundError, *DecodeError, *TTSError, and the rest) on failure. Pass
// that error to ClassifyError to get a coarse-grained ErrorClass together
// with a short, user-facing message describing what went wrong.
//
// The package depends only on the Go standard library.
package hub

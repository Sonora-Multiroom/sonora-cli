package hub

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// ErrorClass classifies a failure into one of the exit-code classes the CLI
// must distinguish (constitution Principle V).
type ErrorClass int

const (
	ClassNone ErrorClass = iota
	ClassUsage
	ClassHub
	ClassNetwork
)

// ExitCode returns the CLI exit code for this error class, per research.md §6.
func (c ErrorClass) ExitCode() int {
	switch c {
	case ClassUsage:
		return 2
	case ClassHub:
		return 3
	case ClassNetwork:
		return 4
	default:
		return 0
	}
}

// StatusError indicates the hub responded with a non-2xx HTTP status.
type StatusError struct {
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("hub returned HTTP %d", e.StatusCode)
}

// DecodeError indicates the hub's response body did not match the expected
// shape (FR-013): missing/extra-typed fields, or a required field empty.
type DecodeError struct {
	Err error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("malformed response from hub: %v", e.Err)
}

func (e *DecodeError) Unwrap() error { return e.Err }

// ClassifyError maps an error from a hub API call to its exit-code class
// and a short, friendly, user-facing message. The underlying error remains
// available to the caller for --verbose output.
func ClassifyError(err error) (class ErrorClass, friendlyMsg string) {
	if err == nil {
		return ClassNone, ""
	}

	var statusErr *StatusError
	if errors.As(err, &statusErr) {
		return ClassHub, fmt.Sprintf("hub reported an error (HTTP %d)", statusErr.StatusCode)
	}

	var decodeErr *DecodeError
	if errors.As(err, &decodeErr) {
		return ClassHub, "hub returned an unexpected or malformed response"
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return ClassNetwork, "hub did not respond in time"
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ClassNetwork, "hub did not respond in time"
		}
		return ClassNetwork, "could not reach the hub"
	}

	return ClassNetwork, "could not reach the hub"
}

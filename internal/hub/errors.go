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
	ClassNotFound
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
	case ClassNotFound:
		return 5
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

// NotFoundError indicates the hub responded 404 for a specific output
// identifier, distinct from a generic StatusError (FR-012).
type NotFoundError struct {
	OutputID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("output not found: %s", e.OutputID)
}

// ClassifyError maps an error from a hub API call to its exit-code class
// and a short, friendly, user-facing message. The underlying error remains
// available to the caller for --verbose output.
func ClassifyError(err error) (class ErrorClass, friendlyMsg string) {
	if err == nil {
		return ClassNone, ""
	}

	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		return ClassNotFound, "output not found: " + notFoundErr.OutputID
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

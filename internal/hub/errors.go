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
	ClassValidation
	ClassRouteFailed
	ClassSourceUnreachable
	ClassServiceUnavailable
	ClassInputNotFound
	ClassTargetNotFound
)

// ExitCode returns the CLI exit code for this error class, per research.md §6.
// Exit code 7 ("target matches both an output and a group") is retired —
// path-style target addressing makes that case structurally unreachable
// (data-model.md) — and is not reused by any other class.
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
	case ClassValidation:
		return 6
	case ClassRouteFailed:
		return 8
	case ClassSourceUnreachable:
		return 9
	case ClassServiceUnavailable:
		return 10
	case ClassInputNotFound:
		return 11
	case ClassTargetNotFound:
		return 12
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

// NotFoundError indicates the hub responded 404 for a specific resource
// identifier (e.g. "output", "input"), distinct from a generic StatusError
// (FR-012).
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// APIError indicates the hub responded with a non-2xx status and a decodable
// #/components/schemas/ErrorResponse body (FR-009): 400/422/502/503 for
// POST /api/v2/play.
type APIError struct {
	StatusCode int
	Title      string
	Detail     string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return fmt.Sprintf("hub returned HTTP %d: %s", e.StatusCode, e.Title)
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
		return ClassNotFound, fmt.Sprintf("%s not found: %s", notFoundErr.Resource, notFoundErr.ID)
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		msg := apiErr.Detail
		if msg == "" {
			msg = apiErr.Title
		}
		switch apiErr.StatusCode {
		case 400:
			if msg == "" {
				msg = "the request was rejected as invalid"
			}
			return ClassValidation, msg
		case 422:
			if msg == "" {
				msg = "route creation failed"
			}
			return ClassRouteFailed, msg
		case 502:
			return ClassSourceUnreachable, "the audio source could not be reached"
		case 503:
			return ClassServiceUnavailable, "the hub's playback service is temporarily unavailable"
		default:
			return ClassHub, fmt.Sprintf("hub reported an error (HTTP %d)", apiErr.StatusCode)
		}
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

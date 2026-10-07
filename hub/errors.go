package hub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

// ErrorClass classifies a failure returned by a hub API call into a coarse
// category, so a caller can decide how to react without inspecting the
// underlying error type directly.
type ErrorClass int

const (
	ClassNone ErrorClass = iota
	ClassHub
	ClassNetwork
	ClassNotFound
	ClassValidation
	ClassRouteFailed
	ClassSourceUnreachable
	ClassServiceUnavailable
	ClassInputNotFound
	ClassTargetNotFound
	ClassTTSUnavailable
	// ClassConflict: the hub refused the request in its current state (409),
	// for example a disabled target; change that state and retry.
	ClassConflict
)

// StatusError indicates the hub responded with a non-2xx HTTP status.
type StatusError struct {
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("hub returned HTTP %d", e.StatusCode)
}

// DecodeError indicates the hub's response body did not match the expected
// shape: missing/extra-typed fields, or a required field empty.
type DecodeError struct {
	Err error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("malformed response from hub: %v", e.Err)
}

func (e *DecodeError) Unwrap() error { return e.Err }

// NotFoundError indicates the hub responded 404 for a specific resource
// identifier (e.g. "output", "input"), distinct from a generic StatusError.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// APIError indicates the hub responded with a non-2xx status and a decodable
// #/components/schemas/ErrorResponse body.
type APIError struct {
	StatusCode int
	Title      string
	Detail     string
	// Reason says why the hub refused a route (409 from CreateRoute,
	// TransferRoute or Playback): one of INPUT_DISABLED, OUTPUT_DISABLED,
	// GROUP_DISABLED, ALL_MEMBERS_DISABLED, ROUTE_LIMIT_REACHED or
	// INPUT_ALREADY_ON_OUTPUT. The hub may add reasons; treat an unknown one
	// as a generic conflict. Empty on every other error.
	Reason string
	// OutputID names the first output that failed an output-level check of a
	// route refusal. Empty for INPUT_DISABLED, GROUP_DISABLED and every other
	// error.
	OutputID string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return fmt.Sprintf("hub returned HTTP %d: %s", e.StatusCode, e.Title)
}

// ClassifyError maps an error from a hub API call to a coarse ErrorClass
// and a short, friendly, user-facing message. The underlying error remains
// available via the standard errors package (errors.As, errors.Unwrap) for
// callers that want to log or display more detail.
func ClassifyError(err error) (class ErrorClass, friendlyMsg string) {
	if err == nil {
		return ClassNone, ""
	}

	// TTS branches: *TTSUnavailableError and *TTSError are checked first,
	// since neither can be produced by any of the branches below.
	var unavailErr *TTSUnavailableError
	if errors.As(err, &unavailErr) {
		if unavailErr.Diagnosis == DiagnosisHubAddress {
			return ClassNetwork, unavailErr.Error()
		}
		return ClassTTSUnavailable, unavailErr.Error()
	}

	var ttsErr *TTSError
	if errors.As(err, &ttsErr) {
		msg := ttsErr.Error()
		switch ttsErr.Code {
		case "TARGET_NOT_FOUND":
			return ClassTargetNotFound, msg
		case "INVALID_REQUEST":
			return ClassValidation, msg
		case "PROVIDER_NOT_FOUND":
			return ClassNotFound, msg
		case "PROVIDER_TIMEOUT", "PROVIDER_RATE_LIMITED", "PROVIDER_ERROR", "FORMAT_NORMALIZATION_FAILED":
			return ClassServiceUnavailable, msg
		}
		switch ttsErr.StatusCode {
		case 400:
			return ClassValidation, msg
		case 503:
			return ClassServiceUnavailable, msg
		default:
			return ClassHub, msg
		}
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
		case 409:
			if msg == "" {
				msg = conflictMessage
			}
			return ClassConflict, msg
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
		if statusErr.StatusCode == 409 {
			return ClassConflict, conflictMessage
		}
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

// conflictMessage is ClassifyError's message for a 409 whose body gives no
// detail or title.
const conflictMessage = "the hub refused the request in its current state; change that state and retry"

// apiErrorFromBody decodes an error response body into an *APIError for the
// given status, falling back to a *StatusError when the body is not a
// decodable ErrorResponse.
func apiErrorFromBody(status int, body io.Reader) error {
	var errBody errorResponse
	if err := json.NewDecoder(body).Decode(&errBody); err != nil {
		return &StatusError{StatusCode: status}
	}
	return &APIError{
		StatusCode: status,
		Title:      errBody.Title,
		Detail:     errBody.Detail,
		Reason:     errBody.Reason,
		OutputID:   errBody.OutputID,
	}
}

// notFoundFromBody returns the *NotFoundError for a 404 from an operation
// whose 404 can mean more than one missing resource (createRoute: input or
// target; transferRoute: route or target). The hub's problem detail names
// the missing one as "<Resource> not found: <id>" (e.g. "Output not found:
// kitchen"); fallback is returned when the body has no such detail.
func notFoundFromBody(body io.Reader, fallback NotFoundError) *NotFoundError {
	var errBody errorResponse
	if err := json.NewDecoder(body).Decode(&errBody); err != nil {
		return &fallback
	}
	resource, id, ok := strings.Cut(errBody.Detail, " not found: ")
	if !ok || resource == "" || id == "" || strings.ContainsAny(resource, " ,") {
		return &fallback
	}
	return &NotFoundError{Resource: strings.ToLower(resource), ID: id}
}

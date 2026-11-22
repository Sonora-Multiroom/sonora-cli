package hub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
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
	// ClassRadioUnavailable: the hub does not serve the radio catalogue API (the
	// radio extension is missing, disabled, failed to load, inactive, or too old
	// to search stations).
	ClassRadioUnavailable
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

// InfoNotOfferedError indicates the hub answered 404 to GET /api/v2/info,
// which is what a hub older than 0.1.33 does: it does not offer hub info.
type InfoNotOfferedError struct{}

func (e *InfoNotOfferedError) Error() string {
	return "the hub does not offer GET /api/v2/info"
}

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
	// GROUP_DISABLED, ALL_MEMBERS_DISABLED, ROUTE_LIMIT_REACHED,
	// INPUT_ALREADY_ON_OUTPUT or ROUTE_CHANGING_OUTPUTS. The last one, from
	// TransferRoute only, is transient (see ReasonRouteChangingOutputs and
	// IsTransientConflict); the others clear only when the state changes. The
	// hub may add reasons; treat an unknown one as a generic conflict. Empty on
	// every other error.
	Reason string
	// OutputID names the first output that failed an output-level check of a
	// route refusal. Empty for INPUT_DISABLED, GROUP_DISABLED,
	// ROUTE_CHANGING_OUTPUTS and every other error.
	OutputID string
	// Type is the problem's "type" member, for example
	// "urn:multiroom:radio:error:invalid-parameter". Empty when the body has none.
	Type string
	// Parameter names the query or path parameter an invalid-parameter problem
	// is about. Empty on every other error.
	Parameter string
	// RetryAfter is the delay the hub's Retry-After header asked for, in whole
	// seconds. Zero when the header is absent or not an integer.
	RetryAfter time.Duration
}

// ReasonRouteChangingOutputs is the APIError.Reason of a transfer refused while the
// route is still adding an output that was just enabled. It clears by itself within
// moments, so retrying the same request shortly is the right response.
const ReasonRouteChangingOutputs = "ROUTE_CHANGING_OUTPUTS"

// The 409 reasons HUB-041 (hub 0.1.35) added for kept inputs. Only
// ReasonInputUnplayable clears without a change by the caller, once the
// input's extension is running.
const (
	// ReasonInputUnplayable: a kept input's link has no handler or resolver, for
	// example because its extension is not running.
	ReasonInputUnplayable = "INPUT_UNPLAYABLE"
	// ReasonInputIDTaken: another input already has the id.
	ReasonInputIDTaken = "INPUT_ID_TAKEN"
	// ReasonInputStoreFull: the store already keeps the most inputs it may.
	ReasonInputStoreFull = "INPUT_STORE_FULL"
)

// IsTransientConflict reports whether err is a 409 route refusal that clears without
// any change of state (today only ReasonRouteChangingOutputs). It is false for every
// other error, including other 409 refusals.
func IsTransientConflict(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 409 && apiErr.Reason == ReasonRouteChangingOutputs
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

	// TTS branches: *TTSUnavailableError, *TTSNotOfferedError and *TTSError
	// are checked first, since none can be produced by any of the branches
	// below.
	var unavailErr *TTSUnavailableError
	if errors.As(err, &unavailErr) {
		if unavailErr.Diagnosis == DiagnosisHubAddress {
			return ClassNetwork, unavailErr.Error()
		}
		return ClassTTSUnavailable, unavailErr.Error()
	}

	var radioUnavail *RadioUnavailableError
	if errors.As(err, &radioUnavail) {
		if radioUnavail.Diagnosis == DiagnosisHubAddress {
			return ClassNetwork, radioUnavail.Error()
		}
		return ClassRadioUnavailable, radioUnavail.Error()
	}

	var radioNotOffered *RadioNotOfferedError
	if errors.As(err, &radioNotOffered) {
		return ClassRadioUnavailable, "radio is not available on this hub"
	}

	var infoNotOffered *InfoNotOfferedError
	if errors.As(err, &infoNotOffered) {
		return ClassHub, "this hub does not offer hub info (needs hub 0.1.33 or later)"
	}

	var notOfferedErr *TTSNotOfferedError
	if errors.As(err, &notOfferedErr) {
		return ClassTTSUnavailable, "text-to-speech is not available on this hub"
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
		case "PROVIDER_TIMEOUT", "PROVIDER_RATE_LIMITED", "PROVIDER_ERROR", "FORMAT_NORMALIZATION_FAILED", "CACHE_WRITE_FAILED":
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
	if errors.As(err, &apiErr) && strings.HasPrefix(apiErr.Type, radioProblemPrefix) {
		return classifyRadioProblem(apiErr)
	}
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
			if apiErr.Reason == ReasonRouteChangingOutputs {
				return ClassConflict, "the route is still adding an output; retry in a moment"
			}
			switch apiErr.Reason {
			case ReasonInputUnplayable:
				return ClassConflict, "the input's link is not playable now, perhaps because its extension is not running; retry once it is"
			case ReasonInputIDTaken:
				return ClassConflict, "an input with that id already exists"
			case ReasonInputStoreFull:
				return ClassConflict, "the hub's store of kept inputs is full; delete a kept input first"
			}
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
			if apiErr.StatusCode == 500 && msg != "" {
				// A kept input's store failure: the detail names the cause.
				return ClassHub, "the hub could not save the input: " + msg
			}
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

// radioProblemPrefix starts the "type" of every problem the radio extension
// answers with.
const radioProblemPrefix = "urn:multiroom:radio:error:"

// classifyRadioProblem maps a radio extension problem onto a class and message.
func classifyRadioProblem(e *APIError) (ErrorClass, string) {
	detail := e.Detail
	if detail == "" {
		detail = e.Title
	}
	switch strings.TrimPrefix(e.Type, radioProblemPrefix) {
	case "invalid-parameter":
		if detail == "" {
			detail = "the request was rejected as invalid"
		}
		return ClassValidation, detail
	case "catalogue-bad-answer":
		msg := "the radio catalogue gave an unusable answer"
		if detail != "" {
			msg += ": " + detail
		}
		return ClassSourceUnreachable, msg
	case "catalogue-unavailable":
		if secs := int64(e.RetryAfter / time.Second); secs > 0 {
			return ClassServiceUnavailable, fmt.Sprintf("the radio catalogue is unavailable; retry in %d s", secs)
		}
		return ClassServiceUnavailable, "the radio catalogue is unavailable; retry later"
	default:
		if detail == "" {
			detail = fmt.Sprintf("hub reported an error (HTTP %d)", e.StatusCode)
		}
		return ClassHub, detail
	}
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

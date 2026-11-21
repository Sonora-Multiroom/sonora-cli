package hub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Extension mirrors #/components/schemas/Extension in api/openapi.json
// field-for-field, decoding only the fields an availability diagnosis needs:
// name, requiredApiVersion, and connectionState are not decoded.
type Extension struct {
	ID string `json:"id"`
	// Version is the extension's own version, nil when the hub does not report one.
	Version         *string `json:"version"`
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejectionReason"`
}

// ExtensionInventory mirrors #/components/schemas/ExtensionInventory in
// api/openapi.json field-for-field. extensionsDirectory is not decoded,
// because nothing uses it.
type ExtensionInventory struct {
	LoadingEnabled *bool       `json:"loadingEnabled"`
	Extensions     []Extension `json:"extensions"`
}

// ListExtensions calls GET {baseURL}/api/v2/extensions (operationId
// "listExtensions") and returns the decoded inventory. Callers use it to
// diagnose why an extension's operation was not offered, by checking that
// extension's status. Any non-2xx status is a *StatusError, and an
// undecodable body is a *DecodeError. A missing extensions array decodes as
// a nil (empty) slice, which range treats identically to an empty one.
func ListExtensions(ctx context.Context, client *http.Client, baseURL string) (*ExtensionInventory, error) {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/extensions"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

	var inv ExtensionInventory
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &inv, nil
}

// ExtensionDiagnosis explains why an extension's API is not served, as
// determined from the hub's extension inventory (ListExtensions) after a
// request to that API was answered 404.
type ExtensionDiagnosis int

const (
	// DiagnosisUnknown: the inventory could not be read, or gives a status this
	// package does not know.
	DiagnosisUnknown ExtensionDiagnosis = iota
	// DiagnosisNotInstalled: the inventory has no entry for the extension.
	DiagnosisNotInstalled
	// DiagnosisDisabled: the extension is installed but switched off.
	DiagnosisDisabled
	// DiagnosisRejected: the hub refused to load the extension.
	DiagnosisRejected
	// DiagnosisInert: the extension is loaded but inactive.
	DiagnosisInert
	// DiagnosisVersionMismatch: the extension is active yet does not serve the
	// operation, so it is older than the operation needs.
	DiagnosisVersionMismatch
	// DiagnosisHubAddress: the inventory itself answered 404, so the base URL
	// is not a Multiroom Audio Hub API.
	DiagnosisHubAddress
)

// extensionFinding is the result of diagnoseExtension.
type extensionFinding struct {
	Diagnosis       ExtensionDiagnosis
	Reason          string
	LoadingDisabled bool
	Version         string
	Cause           error
}

// diagnoseExtension makes one ListExtensions call with ctx and client and
// maps the status of the extension with the given id (never its display
// name) onto a diagnosis. It never fails: an inventory that answers 404
// gives DiagnosisHubAddress, and any other inventory failure gives
// DiagnosisUnknown with that failure as Cause.
func diagnoseExtension(ctx context.Context, client *http.Client, baseURL, id string) extensionFinding {
	inv, err := ListExtensions(ctx, client, baseURL)
	if err != nil {
		var statusErr *StatusError
		if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
			return extensionFinding{Diagnosis: DiagnosisHubAddress}
		}
		return extensionFinding{Diagnosis: DiagnosisUnknown, Cause: err}
	}

	for _, ext := range inv.Extensions {
		if ext.ID != id {
			continue
		}
		switch ext.Status {
		case "DISABLED":
			return extensionFinding{Diagnosis: DiagnosisDisabled}
		case "REJECTED":
			reason := ""
			if ext.RejectionReason != nil {
				reason = SingleLine(*ext.RejectionReason)
			}
			return extensionFinding{Diagnosis: DiagnosisRejected, Reason: reason}
		case "INERT":
			return extensionFinding{Diagnosis: DiagnosisInert}
		case "ACTIVE":
			version := ""
			if ext.Version != nil {
				version = SingleLine(*ext.Version)
			}
			return extensionFinding{Diagnosis: DiagnosisVersionMismatch, Version: version}
		default:
			return extensionFinding{Diagnosis: DiagnosisUnknown}
		}
	}

	loadingDisabled := inv.LoadingEnabled != nil && !*inv.LoadingEnabled
	return extensionFinding{Diagnosis: DiagnosisNotInstalled, LoadingDisabled: loadingDisabled}
}

// RadioUnavailableError is the explained form of RadioNotOfferedError: the
// radio extension's API is not served, and Diagnosis says why. It is built by
// DiagnoseRadioUnavailable, not returned by a radio call.
type RadioUnavailableError struct {
	Diagnosis ExtensionDiagnosis
	// Reason is the hub's rejection reason, on one line (DiagnosisRejected).
	Reason string
	// LoadingDisabled is true when the hub's configuration switches extension
	// loading off (DiagnosisNotInstalled).
	LoadingDisabled bool
	// Version is the installed extension's version when the inventory gives one
	// (DiagnosisVersionMismatch).
	Version string
	// BaseURL is the address that was tried (DiagnosisHubAddress).
	BaseURL string
	// Cause is the inventory failure behind DiagnosisUnknown, when there is one.
	Cause error
}

// Error renders why radio is not available, specific to the diagnosis.
func (e *RadioUnavailableError) Error() string {
	if e.Diagnosis == DiagnosisHubAddress {
		return hubAddressMessage(e.BaseURL)
	}

	const head = "radio is not available on this hub"
	switch e.Diagnosis {
	case DiagnosisNotInstalled:
		if e.LoadingDisabled {
			return head + ": the radio extension is not installed on this hub (extension loading is switched off in the hub's configuration)"
		}
		return head + ": the radio extension is not installed on this hub"
	case DiagnosisDisabled:
		return head + ": the radio extension is installed but disabled in the hub's configuration"
	case DiagnosisRejected:
		if e.Reason != "" {
			return head + ": the radio extension failed to load: " + e.Reason
		}
		return head + ": the radio extension failed to load"
	case DiagnosisInert:
		return head + ": the radio extension is loaded but inactive"
	case DiagnosisVersionMismatch:
		version := "version unknown"
		if e.Version != "" {
			version = "version " + e.Version
		}
		return head + ": the radio extension on this hub (" + version + ") cannot search stations; it needs 0.2.0 or later"
	default: // DiagnosisUnknown
		return head
	}
}

// Unwrap returns the inventory failure behind the diagnosis, if any.
func (e *RadioUnavailableError) Unwrap() error { return e.Cause }

// DiagnoseRadioUnavailable explains why a radio operation returned
// *RadioNotOfferedError. It makes one ListExtensions call with ctx and
// client, so the caller's deadline and timeout bound it, and matches the
// extension by its id "radio", never its display name. It never returns nil.
func DiagnoseRadioUnavailable(ctx context.Context, client *http.Client, baseURL string) *RadioUnavailableError {
	f := diagnoseExtension(ctx, client, baseURL, "radio")
	e := &RadioUnavailableError{
		Diagnosis: f.Diagnosis, Reason: f.Reason, LoadingDisabled: f.LoadingDisabled,
		Version: f.Version, Cause: f.Cause,
	}
	if f.Diagnosis == DiagnosisHubAddress {
		e.BaseURL = baseURL
	}
	return e
}

// hubAddressMessage is the text for DiagnosisHubAddress, shared by every
// extension's unavailable error.
func hubAddressMessage(baseURL string) string {
	return fmt.Sprintf("%s is not serving the Multiroom Audio Hub API: the hub URL is wrong, or the hub's control API (REST) extension is not installed or not loaded", baseURL)
}

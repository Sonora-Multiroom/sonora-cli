package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Response shapes here mirror the 409 responses of the createRoute,
// transferRoute and playback operations, and the reason/outputId members of
// #/components/schemas/ErrorResponse, in api/openapi.json (constitution
// Principle II): a route the hub refuses in its current state.

// routeRefusingCalls runs each operation that can answer a route refusal
// against baseURL and returns its error.
var routeRefusingCalls = map[string]func(baseURL string) error{
	"CreateRoute": func(baseURL string) error {
		req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "bathroom", TargetType: "SINGLE_OUTPUT"}
		_, err := hub.CreateRoute(context.Background(), hub.NewClient(), baseURL, req)
		return err
	},
	"TransferRoute": func(baseURL string) error {
		req := hub.TransferRequest{TargetID: "bathroom", TargetType: "SINGLE_OUTPUT"}
		_, err := hub.TransferRoute(context.Background(), hub.NewClient(), baseURL, "route_abc123", req)
		return err
	},
	"Playback": func(baseURL string) error {
		req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "bathroom", TargetType: "SINGLE_OUTPUT"}
		_, err := hub.Playback(context.Background(), hub.NewClient(), baseURL, req)
		return err
	},
}

// admissionProblem is the hub's 409 body for a refused route; outputID ""
// omits the member, as the hub does for input- and group-level reasons.
func admissionProblem(reason, outputID, detail string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{
			"type": "urn:multiroom:error:route-admission", "title": "Route not admitted",
			"detail": detail, "status": 409, "reason": reason,
		}
		if outputID != "" {
			body["outputId"] = outputID
		}
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(body)
	}
}

func TestRouteRefusal_409_DecodesReasonAndOutputID(t *testing.T) {
	refusals := []struct {
		reason, outputID, detail string
	}{
		{"OUTPUT_DISABLED", "bathroom", "Output 'bathroom' is disabled"},
		{"ALL_MEMBERS_DISABLED", "bathroom", "Every member of group 'upstairs' is disabled"},
		{"ROUTE_LIMIT_REACHED", "bathroom", "Output 'bathroom' is at its route limit"},
		{"INPUT_ALREADY_ON_OUTPUT", "bathroom", "Input 'spotify-1' already plays on 'bathroom'"},
		{"INPUT_DISABLED", "", "Input 'spotify-1' is disabled"},
		{"GROUP_DISABLED", "", "Group 'upstairs' is disabled"},
		{"SOME_FUTURE_REASON", "", "Refused for a reason this client does not know"},
	}
	for name, call := range routeRefusingCalls {
		for _, tt := range refusals {
			t.Run(name+"/"+tt.reason, func(t *testing.T) {
				srv := httptest.NewServer(admissionProblem(tt.reason, tt.outputID, tt.detail))
				defer srv.Close()

				err := call(srv.URL)
				apiErr, ok := errors.AsType[*hub.APIError](err)
				if !ok {
					t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
				}
				want := hub.APIError{
					StatusCode: http.StatusConflict, Title: "Route not admitted", Detail: tt.detail,
					Reason: tt.reason, OutputID: tt.outputID,
				}
				if *apiErr != want {
					t.Errorf("APIError = %+v, want %+v", *apiErr, want)
				}
				class, msg := hub.ClassifyError(err)
				if class != hub.ClassConflict || msg != tt.detail {
					t.Errorf("ClassifyError = (%v, %q), want (ClassConflict, %q)", class, msg, tt.detail)
				}
			})
		}
	}
}

func TestRouteRefusal_409_UndecodableBodyFallsBackToStatusError(t *testing.T) {
	for name, call := range routeRefusingCalls {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte("not json"))
			}))
			defer srv.Close()

			err := call(srv.URL)
			statusErr, ok := errors.AsType[*hub.StatusError](err)
			if !ok || statusErr.StatusCode != http.StatusConflict {
				t.Fatalf("expected a *hub.StatusError with 409, got %T: %v", err, err)
			}
			if class, _ := hub.ClassifyError(err); class != hub.ClassConflict {
				t.Errorf("ClassifyError class = %v, want ClassConflict", class)
			}
		})
	}
}

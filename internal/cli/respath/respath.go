// Package respath resolves a "resource" or "resource/id" command-line
// argument into a ResourceKind and optional identifier, shared by the
// get/list dispatcher and play's target argument (research.md §1).
package respath

import (
	"fmt"
	"regexp"
	"strings"
)

// ResourceKind is a closed enum of the resources get/list and play address.
type ResourceKind int

const (
	unknownKind ResourceKind = iota
	Inputs
	Outputs
	Groups
	Routes
)

// String returns the canonical resource name.
func (k ResourceKind) String() string {
	switch k {
	case Inputs:
		return "inputs"
	case Outputs:
		return "outputs"
	case Groups:
		return "groups"
	case Routes:
		return "routes"
	default:
		return "unknown"
	}
}

// canonicalOrder fixes the enumeration order used by Names() and in
// FR-006a's "missing resource" error message.
var canonicalOrder = []ResourceKind{Inputs, Outputs, Groups, Routes}

var names = map[string]ResourceKind{
	"inputs": Inputs, "in": Inputs,
	"outputs": Outputs, "out": Outputs,
	"groups": Groups, "gr": Groups,
	"routes": Routes, "rt": Routes,
}

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,255}$`)

// Path is the parsed form of a "resource" or "resource/id" argument. ID is
// empty for the collection form.
type Path struct {
	Kind ResourceKind
	ID   string
}

// Parse splits arg on the first '/' into a resource-name segment and an
// optional identifier segment (FR-004a). The resource-name segment must
// name a known resource (or alias, once registered); the identifier
// segment, if present, must match ^[a-zA-Z0-9_-]{1,255}$ — a second '/'
// inside it therefore fails id validation rather than being treated as a
// deeper path.
func Parse(arg string) (Path, error) {
	name, id, hasID := arg, "", false
	if idx := strings.IndexByte(arg, '/'); idx >= 0 {
		name, id, hasID = arg[:idx], arg[idx+1:], true
	}

	kind, ok := names[name]
	if !ok {
		return Path{}, fmt.Errorf("unrecognized resource %q", name)
	}
	if hasID {
		if !idPattern.MatchString(id) {
			return Path{}, fmt.Errorf("malformed identifier %q in %q", id, arg)
		}
		return Path{Kind: kind, ID: id}, nil
	}
	return Path{Kind: kind}, nil
}

// Names returns the canonical resource names, in a fixed order, for use in
// usage/error messages (FR-006a).
func Names() []string {
	out := make([]string, len(canonicalOrder))
	for i, k := range canonicalOrder {
		out[i] = k.String()
	}
	return out
}

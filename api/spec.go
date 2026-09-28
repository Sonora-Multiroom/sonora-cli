// Package api publishes the Multiroom Audio Hub OpenAPI specification that
// the hub package at this same module version is tested against.
package api

import _ "embed"

// Spec is the exact contents of the hub's OpenAPI specification document.
// Callers must not modify this slice.
//
//go:embed openapi.json
var Spec []byte

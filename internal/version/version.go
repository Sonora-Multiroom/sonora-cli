// Package version holds the sonora CLI's version string.
package version

// Version is set at build time via -ldflags (see build.sh). It falls back
// to "dev" for a plain `go build`/`go run` without that flag.
var Version = "dev"

// Package version exposes build metadata baked in via -ldflags.
package version

import "runtime"

var (
	// Version is the semantic version of the build.
	Version = "0.0.0-dev"
	// Commit is the git SHA of the build.
	Commit = "unknown"
	// BuildDate is the RFC3339 build timestamp.
	BuildDate = "unknown"
)

// GoVersion returns the Go runtime version used to build the binary.
func GoVersion() string {
	return runtime.Version()
}

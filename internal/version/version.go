// Package version holds build-time version information.
//
// Values are injected via -ldflags at build time (Makefile, GoReleaser):
//
//	go build -ldflags "-X github.com/appsecomega/veracode-go-cli/internal/version.Version=v0.1.0 ..."
//
// When built via `go install pkg@version` (no ldflags), Version falls back
// to the module version embedded by the Go toolchain.
package version

import "runtime/debug"

// Defaults are overridden at release time. Keep "dev" as fallback
// so `go run` / local `go build` without ldflags still works.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// EffectiveVersion returns the ldflags-injected version when present,
// otherwise the module version stamped by `go install pkg@version`.
// Local builds (`go run`, `go build` without ldflags) report "(devel)",
// which we normalize back to "dev".
func EffectiveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return Version
}

// Short returns a human-readable one-line version string.
func Short() string {
	return "veracode-go-cli " + EffectiveVersion() + " (commit " + Commit + ", built " + Date + ")"
}

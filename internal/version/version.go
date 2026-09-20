// Package version holds build-time version information.
//
// Values are injected via -ldflags at build time (Makefile, GoReleaser):
//
//	go build -ldflags "-X github.com/appsecomega/veracode-go-cli/internal/version.Version=v0.1.0 ..."
package version

// Defaults are overridden at release time. Keep "dev" as fallback
// so `go install` / `go run` without ldflags still works.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Short returns a human-readable one-line version string.
func Short() string {
	return "veracode-go-cli " + Version + " (commit " + Commit + ", built " + Date + ")"
}

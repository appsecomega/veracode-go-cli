// Command veracode-go-cli is the entry point of the CLI.
// All logic lives in internal/cli so main stays trivial and
// `go install` works from this package.
package main

import (
	"fmt"
	"os"

	"github.com/appsecomega/veracode-go-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

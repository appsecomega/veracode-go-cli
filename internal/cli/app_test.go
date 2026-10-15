package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/appsecomega/veracode-go-cli/internal/veracode"
)

func testApp() *veracode.Application {
	app := &veracode.Application{
		ID:                    123456,
		GUID:                  "a99686b7-b54a-4fc4-abf5-cac0315fc041",
		AppProfileURL:         "HomeAppProfile:72529:123456",
		ResultsURL:            "ViewReportsResultSummary:72529:123456:21801526",
		LastCompletedScanDate: "2022-10-31T13:41:35.000Z",
	}
	app.Profile.Name = "WebGoat-Legacy-master"
	return app
}

func TestPrintBareURLs(t *testing.T) {
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)

	if err := printBareURLs(cmd, testApp(), true, true); err != nil {
		t.Fatalf("printBareURLs: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), out.String())
	}
	if !strings.HasPrefix(lines[0], "https://analysiscenter.veracode.com/auth/index.jsp#") {
		t.Errorf("app line = %q", lines[0])
	}
	if !strings.Contains(lines[1], "ViewReportsResultSummary:") {
		t.Errorf("scan line = %q", lines[1])
	}
}

func TestPrintBareURLsWithoutPublishedScan(t *testing.T) {
	app := testApp()
	app.ResultsURL = ""

	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)

	if err := printBareURLs(cmd, app, false, true); err == nil {
		t.Fatal("expected error when no scan published, got nil")
	}
}

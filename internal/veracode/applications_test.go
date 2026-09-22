package veracode

import (
	"encoding/json"
	"strings"
	"testing"
)

// Test string padrão do projeto (não faz chamada de rede).
const testAppName = "WebGoat-Legacy-master"

func TestBuildAnalysisCenterURL(t *testing.T) {
	got := BuildAnalysisCenterURL("/some/profile/url")
	wantPrefix := "https://analysiscenter.veracode.com/auth/index.jsp#"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("BuildAnalysisCenterURL = %q, want prefix %q", got, wantPrefix)
	}
	if !strings.HasSuffix(got, "/some/profile/url") {
		t.Fatalf("BuildAnalysisCenterURL = %q, want suffix preserved", got)
	}
}

func TestRemoveCredentialPrefix(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"abc123-SECRETHEX", "SECRETHEX"},
		{"SECRETHEX", "SECRETHEX"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := removeCredentialPrefix(tt.in); got != tt.want {
			t.Errorf("removeCredentialPrefix(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTestAppNameIsNotEmpty(t *testing.T) {
	if strings.TrimSpace(testAppName) == "" {
		t.Fatal("test app name must not be empty")
	}
}

func TestBuildLatestScanResultsURL(t *testing.T) {
	got := BuildLatestScanResultsURL("ViewReportsResultSummary:72529:1571531:21801526")
	want := "https://analysiscenter.veracode.com/auth/index.jsp#ViewReportsResultSummary:72529:1571531:21801526"
	if got != want {
		t.Fatalf("BuildLatestScanResultsURL = %q, want %q", got, want)
	}

	if got := BuildLatestScanResultsURL(""); got != "" {
		t.Fatalf("BuildLatestScanResultsURL(\"\") = %q, want empty", got)
	}
}

func TestParseApplicationWithLatestScan(t *testing.T) {
	// Realistic GET /appsec/v1/applications?name=... payload (fields trimmed).
	body := `{"_embedded":{"applications":[{` +
		`"id":123456,` +
		`"guid":"a99686b7-b54a-4fc4-abf5-cac0315fc041",` +
		`"app_profile_url":"HomeAppProfile:72529:123456",` +
		`"results_url":"ViewReportsResultSummary:72529:123456:21801526",` +
		`"last_completed_scan_date":"2022-10-31T13:41:35.000Z",` +
		`"scans":[{"scan_type":"STATIC","status":"PUBLISHED","date":"2022-10-31T13:41:35.000Z"}],` +
		`"profile":{"name":"WebGoat-Legacy-master"},` +
		`"_links":{"self":{"href":"https://api.veracode.com/appsec/v1/applications/a99686b7-b54a-4fc4-abf5-cac0315fc041"}}` +
		`}]}}`

	var result ApplicationsResponse
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if len(result.Embedded.Applications) != 1 {
		t.Fatalf("got %d applications, want 1", len(result.Embedded.Applications))
	}

	app := result.Embedded.Applications[0]

	if app.Profile.Name != testAppName {
		t.Errorf("name = %q, want %q", app.Profile.Name, testAppName)
	}
	if app.LastCompletedScanDate == "" {
		t.Error("last_completed_scan_date was not parsed")
	}
	if len(app.Scans) != 1 || app.Scans[0].ScanType != "STATIC" {
		t.Errorf("scans = %+v, want one STATIC scan", app.Scans)
	}

	link := BuildLatestScanResultsURL(app.ResultsURL)
	if !strings.HasPrefix(link, "https://analysiscenter.veracode.com/auth/index.jsp#ViewReportsResultSummary:") {
		t.Errorf("latest scan link = %q, want ViewReportsResultSummary deep link", link)
	}
}

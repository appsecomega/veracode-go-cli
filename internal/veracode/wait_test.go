package veracode

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// waitFixture serves both the REST applications endpoint and the XML
// buildinfo endpoint with scripted responses.
type waitFixture struct {
	t *testing.T

	buildinfoCalls int
	// buildinfoScript returns the XML body for the n-th buildinfo call
	// (1-based). Returning "" keeps the previous behavior.
	buildinfoScript func(call int) string
	// appScript returns the REST body for the n-th applications call.
	appScript func(call int) string
	appCalls  int
}

func (f *waitFixture) handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/appsec/v1/applications":
		f.appCalls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, f.appScript(f.appCalls))
	case "/api/5.0/getbuildinfo.do":
		f.buildinfoCalls++
		w.Header().Set("Content-Type", "application/xml")
		_, _ = fmt.Fprint(w, f.buildinfoScript(f.buildinfoCalls))
	default:
		f.t.Errorf("unexpected request path %q", r.URL.Path)
		http.NotFound(w, r)
	}
}

func appJSON(resultsURL string) string {
	return `{"_embedded":{"applications":[{` +
		`"id":123456,` +
		`"guid":"a99686b7-b54a-4fc4-abf5-cac0315fc041",` +
		`"app_profile_url":"HomeAppProfile:72529:123456",` +
		`"results_url":"` + resultsURL + `",` +
		`"last_completed_scan_date":"2022-10-31T13:41:35.000Z",` +
		`"profile":{"name":"WebGoat-Legacy-master"},` +
		`"_links":{"self":{"href":"https://api.veracode.com/appsec/v1/applications/a99686b7-b54a-4fc4-abf5-cac0315fc041"}}` +
		`}]}}`
}

func buildinfoXML(resultsReady bool, status string) string {
	ready := "false"
	if resultsReady {
		ready = "true"
	}
	return `<buildinfo buildinfo_version="1.5" account_id="72529" app_id="123456" build_id="21801526">` +
		`<build build_id="21801526" version="WebGoat-Legacy-master Static" results_ready="` + ready + `">` +
		`<analysis_unit analysis_type="Static" status="` + status + `"/>` +
		`</build></buildinfo>`
}

func setupWaitServer(t *testing.T, f *waitFixture) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(f.handler))
	t.Cleanup(server.Close)

	oldAPI, oldXML := apiBaseURL, xmlAPIBaseURL
	apiBaseURL, xmlAPIBaseURL = server.URL, server.URL
	t.Cleanup(func() { apiBaseURL, xmlAPIBaseURL = oldAPI, oldXML })

	return NewClient(newTestCredentials())
}

func noopSleep(time.Duration) {}

func TestWaitReturnsImmediatelyWhenAlreadyPublished(t *testing.T) {
	f := &waitFixture{t: t,
		buildinfoScript: func(int) string { return buildinfoXML(true, "Results Ready") },
		appScript:       func(int) string { return appJSON("ViewReportsResultSummary:72529:123456:21801526") },
	}
	client := setupWaitServer(t, f)

	app, err := client.WaitForLatestScanResults("WebGoat-Legacy-master", WaitOptions{
		Interval: time.Millisecond,
		Timeout:  time.Second,
		Sleep:    noopSleep,
	})
	if err != nil {
		t.Fatalf("WaitForLatestScanResults: %v", err)
	}
	if app.Profile.Name != "WebGoat-Legacy-master" {
		t.Errorf("name = %q", app.Profile.Name)
	}
	if f.buildinfoCalls != 1 {
		t.Errorf("buildinfo calls = %d, want exactly 1 (no polling)", f.buildinfoCalls)
	}
}

func TestWaitBlocksUntilNewResultsPublished(t *testing.T) {
	const oldURL = "ViewReportsResultSummary:72529:123456:11111111"
	const newURL = "ViewReportsResultSummary:72529:123456:21801526"

	f := &waitFixture{t: t}
	f.buildinfoScript = func(call int) string {
		if call < 3 {
			return buildinfoXML(false, "Scan In Process")
		}
		return buildinfoXML(true, "Results Ready")
	}
	f.appScript = func(call int) string {
		if call < 3 {
			return appJSON(oldURL)
		}
		return appJSON(newURL)
	}
	client := setupWaitServer(t, f)

	app, err := client.WaitForLatestScanResults("WebGoat-Legacy-master", WaitOptions{
		Interval: time.Millisecond,
		Timeout:  5 * time.Second,
		Sleep:    noopSleep,
	})
	if err != nil {
		t.Fatalf("WaitForLatestScanResults: %v", err)
	}
	if app.ResultsURL != newURL {
		t.Errorf("ResultsURL = %q, want %q", app.ResultsURL, newURL)
	}
	if f.buildinfoCalls < 2 {
		t.Errorf("buildinfo calls = %d, want polling to have happened", f.buildinfoCalls)
	}
}

func TestWaitFailsFastOnTerminalStatus(t *testing.T) {
	f := &waitFixture{t: t,
		buildinfoScript: func(int) string { return buildinfoXML(false, "Scan Errors") },
		appScript:       func(int) string { return appJSON("") },
	}
	client := setupWaitServer(t, f)

	_, err := client.WaitForLatestScanResults("WebGoat-Legacy-master", WaitOptions{
		Interval: time.Millisecond,
		Timeout:  5 * time.Second,
		Sleep:    noopSleep,
	})
	if err == nil {
		t.Fatal("expected error for terminal scan status, got nil")
	}
	if !strings.Contains(err.Error(), "Scan Errors") {
		t.Errorf("error = %q, want it to mention the scan status", err.Error())
	}
}

func TestWaitTimesOutInsteadOfHanging(t *testing.T) {
	f := &waitFixture{t: t,
		buildinfoScript: func(int) string { return buildinfoXML(false, "Scan In Process") },
		appScript:       func(int) string { return appJSON("") },
	}
	client := setupWaitServer(t, f)

	start := time.Now()
	_, err := client.WaitForLatestScanResults("WebGoat-Legacy-master", WaitOptions{
		Interval: 10 * time.Millisecond,
		Timeout:  120 * time.Millisecond,
		// Real sleep here: proves wall-clock boundedness.
	})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %q, want a timeout error", err.Error())
	}
	if elapsed > 10*time.Second {
		t.Errorf("wait took %s, expected it bounded near the timeout", elapsed)
	}
}

func TestWaitReturnsImmediatelyWhenNoBuilds(t *testing.T) {
	f := &waitFixture{t: t,
		buildinfoScript: func(int) string { return `<buildinfo buildinfo_version="1.5"/>` },
		appScript:       func(int) string { return appJSON("") },
	}
	client := setupWaitServer(t, f)

	app, err := client.WaitForLatestScanResults("WebGoat-Legacy-master", WaitOptions{
		Interval: time.Millisecond,
		Timeout:  time.Second,
		Sleep:    noopSleep,
	})
	if err != nil {
		t.Fatalf("WaitForLatestScanResults: %v", err)
	}
	if app == nil {
		t.Fatal("expected application, got nil")
	}
}

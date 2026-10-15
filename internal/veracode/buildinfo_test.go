package veracode

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBuildInfoParsesXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/5.0/getbuildinfo.do" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("app_id") != "123456" {
			t.Errorf("unexpected app_id %q", r.URL.Query().Get("app_id"))
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(
			`<buildinfo buildinfo_version="1.5" account_id="72529" app_id="123456" build_id="21801526">` +
				`<build build_id="21801526" version="WebGoat-Legacy-master Static" results_ready="true">` +
				`<analysis_unit analysis_type="Static" status="Results Ready"/>` +
				`</build></buildinfo>`,
		))
	}))
	defer server.Close()

	oldXML := xmlAPIBaseURL
	xmlAPIBaseURL = server.URL
	defer func() { xmlAPIBaseURL = oldXML }()

	client := NewClient(newTestCredentials())

	info, err := client.GetBuildInfo(123456)
	if err != nil {
		t.Fatalf("GetBuildInfo: %v", err)
	}

	if info.AppID != "123456" {
		t.Errorf("AppID = %q, want 123456", info.AppID)
	}
	if !info.Build.ResultsReady {
		t.Error("ResultsReady = false, want true")
	}
	if got := info.StaticStatus(); got != "Results Ready" {
		t.Errorf("StaticStatus = %q, want %q", got, "Results Ready")
	}
}

func TestScanStatusClassification(t *testing.T) {
	inProgress := []string{
		"Scan In Process",
		"Submitted to Engine",
		"Not Submitted to Engine",
		"Pre-Scan Submitted",
		"Pre-Scan Success",
		"Pending Internal Review",
		"Scan On Hold",
		"  scan in process  ",
	}
	for _, s := range inProgress {
		if !ScanInProgress(s) {
			t.Errorf("ScanInProgress(%q) = false, want true", s)
		}
		if ScanTerminalFailure(s) {
			t.Errorf("ScanTerminalFailure(%q) = true, want false", s)
		}
	}

	failures := []string{
		"Incomplete",
		"Scan Errors",
		"Scan Cancelled",
		"Scan Internal Error",
		"Pre-Scan Failed",
		"Pre-Scan Cancelled",
		"No Modules Defined",
	}
	for _, s := range failures {
		if !ScanTerminalFailure(s) {
			t.Errorf("ScanTerminalFailure(%q) = false, want true", s)
		}
		if ScanInProgress(s) {
			t.Errorf("ScanInProgress(%q) = true, want false", s)
		}
	}

	for _, s := range []string{"Results Ready", "Pending Vendor Confirmation", "something-new"} {
		if ScanInProgress(s) {
			t.Errorf("ScanInProgress(%q) = true, want false", s)
		}
		if ScanTerminalFailure(s) {
			t.Errorf("ScanTerminalFailure(%q) = true, want false", s)
		}
	}
}

func TestStaticStatusPrefersStaticUnit(t *testing.T) {
	var info BuildInfo
	info.Build.AnalysisUnits = []struct {
		AnalysisType string `xml:"analysis_type,attr"`
		Status       string `xml:"status,attr"`
	}{
		{AnalysisType: "SCA", Status: "Results Ready"},
		{AnalysisType: "Static", Status: "Scan In Process"},
	}

	if got := info.StaticStatus(); got != "Scan In Process" {
		t.Errorf("StaticStatus = %q, want the Static unit status", got)
	}

	empty := &BuildInfo{}
	if got := empty.StaticStatus(); got != "" {
		t.Errorf("StaticStatus with no units = %q, want empty", got)
	}
}

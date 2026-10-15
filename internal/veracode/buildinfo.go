package veracode

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// BuildInfo mirrors GET /api/5.0/getbuildinfo.do (XML Results API) called
// with app_id and no build_id, which returns the most recent static scan.
//
// Docs: https://docs.veracode.com/r/r_getbuildinfo
type BuildInfo struct {
	XMLName xml.Name `xml:"buildinfo"`
	AppID   string   `xml:"app_id,attr"`
	BuildID string   `xml:"build_id,attr"`
	Build   struct {
		BuildID       string `xml:"build_id,attr"`
		Version       string `xml:"version,attr"`
		ResultsReady  bool   `xml:"results_ready,attr"`
		AnalysisUnits []struct {
			AnalysisType string `xml:"analysis_type,attr"`
			Status       string `xml:"status,attr"`
		} `xml:"analysis_unit"`
	} `xml:"build"`
}

// Documented analysis_unit statuses (buildinfo.xsd StatusType) that mean a
// scan is alive and results may still be published, hence worth waiting for.
var waitableStatuses = map[string]bool{
	"not submitted to engine": true,
	"submitted to engine":     true,
	"pre-scan submitted":      true,
	"pre-scan success":        true,
	"scan in process":         true,
	"pending internal review": true,
	"scan on hold":            true,
}

// Documented statuses that mean the latest build will never publish new
// results (failed/cancelled/invalid scans).
var terminalFailureStatuses = map[string]bool{
	"incomplete":          true,
	"scan errors":         true,
	"scan cancelled":      true,
	"scan internal error": true,
	"pre-scan failed":     true,
	"pre-scan cancelled":  true,
	"no modules defined":  true,
}

func normalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

// StaticStatus returns the status of the Static analysis unit (the driver of
// policy scans). It falls back to the first unit when there is no Static one,
// and to "" when the build carries no units at all.
func (b *BuildInfo) StaticStatus() string {
	for _, unit := range b.Build.AnalysisUnits {
		if strings.EqualFold(strings.TrimSpace(unit.AnalysisType), "Static") {
			return strings.TrimSpace(unit.Status)
		}
	}
	if len(b.Build.AnalysisUnits) > 0 {
		return strings.TrimSpace(b.Build.AnalysisUnits[0].Status)
	}
	return ""
}

// ScanInProgress reports whether status is a documented state in which a scan
// is still alive (running, queued or under review).
func ScanInProgress(status string) bool {
	return waitableStatuses[normalizeStatus(status)]
}

// ScanTerminalFailure reports whether status is a documented terminal state
// after which no new results will be published for this build.
func ScanTerminalFailure(status string) bool {
	return terminalFailureStatuses[normalizeStatus(status)]
}

// GetBuildInfo returns the most recent static scan of an application.
// An empty AppID means the application has no builds at all.
func (c *Client) GetBuildInfo(appID int) (*BuildInfo, error) {
	endpoint, err := url.Parse(xmlAPIBaseURL + "/api/5.0/getbuildinfo.do")
	if err != nil {
		return nil, fmt.Errorf("build API URL: %w", err)
	}

	query := endpoint.Query()
	query.Set("app_id", strconv.Itoa(appID))
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}

	body, err := c.doXML(request)
	if err != nil {
		return nil, err
	}

	var info BuildInfo
	if err := xml.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse API response: %w", err)
	}

	return &info, nil
}

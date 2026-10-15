package veracode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const analysisCenterBaseURL = "https://analysiscenter.veracode.com/auth/index.jsp#"

// GetApplication looks up exactly one application by exact name.
// It errors when zero or more than one application matches.
func (c *Client) GetApplication(name string) (*Application, error) {
	endpoint, err := url.Parse(apiBaseURL + "/appsec/v1/applications")
	if err != nil {
		return nil, fmt.Errorf("build API URL: %w", err)
	}

	query := endpoint.Query()
	query.Set("name", name)
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}

	body, err := c.do(request)
	if err != nil {
		return nil, err
	}

	var result ApplicationsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse API response: %w", err)
	}

	applications := result.Embedded.Applications

	if len(applications) == 0 {
		return nil, fmt.Errorf("application %q was not found", name)
	}
	if len(applications) > 1 {
		return nil, fmt.Errorf(
			"application name %q matched %d applications",
			name,
			len(applications),
		)
	}

	return &applications[0], nil
}

// listPageSize is the page size used when listing all applications
// (API max is 500).
const listPageSize = 100

// ListApplications returns every application in the portfolio, following
// the API pagination automatically.
func (c *Client) ListApplications() ([]Application, error) {
	var all []Application

	for page := 0; ; page++ {
		endpoint, err := url.Parse(apiBaseURL + "/appsec/v1/applications")
		if err != nil {
			return nil, fmt.Errorf("build API URL: %w", err)
		}

		query := endpoint.Query()
		query.Set("page", strconv.Itoa(page))
		query.Set("size", strconv.Itoa(listPageSize))
		endpoint.RawQuery = query.Encode()

		request, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create HTTP request: %w", err)
		}

		body, err := c.do(request)
		if err != nil {
			return nil, err
		}

		var result ApplicationsResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parse API response: %w", err)
		}

		all = append(all, result.Embedded.Applications...)

		if len(result.Embedded.Applications) < listPageSize {
			break
		}
	}

	return all, nil
}

// BuildAnalysisCenterURL builds the deep link to the app profile
// in the Analysis Center UI.
func BuildAnalysisCenterURL(appProfileURL string) string {
	return analysisCenterBaseURL + appProfileURL
}

// BuildLatestScanResultsURL builds the clickable deep link to the latest
// published policy scan (Results page) in the Analysis Center UI.
//
// The API returns results_url as a fragment like
// "ViewReportsResultSummary:<oid>:<app_id>:<build_id>".
// It returns "" when the application has no published policy scan yet.
func BuildLatestScanResultsURL(resultsURL string) string {
	if resultsURL == "" {
		return ""
	}
	return analysisCenterBaseURL + resultsURL
}

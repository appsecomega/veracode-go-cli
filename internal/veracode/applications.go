package veracode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const analysisCenterBaseURL = "https://analysiscenter.veracode.com/auth/index.jsp#"

// GetApplication looks up exactly one application by exact name.
// It errors when zero or more than one application matches.
func (c *Client) GetApplication(name string) (*Application, error) {
	endpoint, err := url.Parse(baseURL + "/appsec/v1/applications")
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

// BuildAnalysisCenterURL builds the deep link to the app profile
// in the Analysis Center UI.
func BuildAnalysisCenterURL(appProfileURL string) string {
	return analysisCenterBaseURL + appProfileURL
}

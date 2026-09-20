// Package veracode implements a minimal client for the Veracode
// AppSec REST APIs (HMAC authentication + application lookup).
package veracode

// ApplicationsResponse mirrors GET /appsec/v1/applications.
type ApplicationsResponse struct {
	Embedded struct {
		Applications []Application `json:"applications"`
	} `json:"_embedded"`
}

// Application is the subset of fields used by this CLI.
type Application struct {
	ID            int    `json:"id"`
	GUID          string `json:"guid"`
	AppProfileURL string `json:"app_profile_url"`

	Profile struct {
		Name string `json:"name"`
	} `json:"profile"`

	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

package veracode

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/appsecomega/veracode-go-cli/internal/config"
)

const baseURL = "https://api.veracode.com"

// Client is a thin wrapper around http.Client with Veracode credentials.
type Client struct {
	Credentials config.Credentials
	HTTPClient  *http.Client
}

// NewClient creates a Client with a sane default timeout.
func NewClient(credentials config.Credentials) *Client {
	return &Client{
		Credentials: credentials,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do executes an authenticated request and returns the raw body.
func (c *Client) do(request *http.Request) ([]byte, error) {
	authHeader, err := generateHMACHeader(c.Credentials.ID, c.Credentials.Secret, request)
	if err != nil {
		return nil, fmt.Errorf("generate HMAC authorization: %w", err)
	}

	request.Header.Set("Authorization", authHeader)
	request.Header.Set("Accept", "application/json")

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute HTTP request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read API response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"Veracode API returned HTTP %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	return body, nil
}

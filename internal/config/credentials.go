// Package config loads credentials and runtime configuration
// from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	// Env vars follow the official Veracode convention.
	APIIDEnv     = "VERACODE_API_KEY_ID"
	APISecretEnv = "VERACODE_API_KEY_SECRET"
)

// Credentials holds a Veracode API id/secret pair.
type Credentials struct {
	ID     string
	Secret string
}

// Load reads credentials from the environment.
// It returns a descriptive error when a variable is missing.
func Load() (Credentials, error) {
	id := strings.TrimSpace(os.Getenv(APIIDEnv))
	secret := strings.TrimSpace(os.Getenv(APISecretEnv))

	if id == "" {
		return Credentials{}, fmt.Errorf("environment variable %s is not set", APIIDEnv)
	}
	if secret == "" {
		return Credentials{}, fmt.Errorf("environment variable %s is not set", APISecretEnv)
	}

	return Credentials{ID: id, Secret: secret}, nil
}

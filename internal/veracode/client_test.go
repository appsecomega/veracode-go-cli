package veracode

import (
	"github.com/appsecomega/veracode-go-cli/internal/config"
)

func newTestCredentials() config.Credentials {
	return config.Credentials{
		ID:     "test-id-00112233",
		Secret: "test-00112233445566778899aabbccddeeff",
	}
}

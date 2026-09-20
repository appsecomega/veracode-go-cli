package veracode

import (
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

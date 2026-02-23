package auth

import (
	"strings"
	"testing"

	"github.com/xokvictor/whoop-mcp/pkg/whoop"
)

func TestGenerateState(t *testing.T) {
	state1, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}
	if state1 == "" {
		t.Error("generateState() returned empty string")
	}

	// Verify uniqueness
	state2, _ := generateState()
	if state1 == state2 {
		t.Error("generateState() should return unique values")
	}

	// Verify it's base64 encoded (no padding issues)
	if len(state1) < 40 {
		t.Errorf("state should be at least 40 chars (32 bytes base64), got %d", len(state1))
	}
}

func TestBuildAuthURL(t *testing.T) {
	config := OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
	}

	url := buildAuthURL(config, "test-state")

	// Should start with WHOOP auth URL
	if !strings.HasPrefix(url, whoop.AuthURL) {
		t.Errorf("URL should start with %s, got %s", whoop.AuthURL, url)
	}

	// Should contain required parameters
	requiredParams := []string{
		"client_id=test-client-id",
		"redirect_uri=",
		"response_type=code",
		"scope=",
		"state=test-state",
	}

	for _, param := range requiredParams {
		if !strings.Contains(url, param) {
			t.Errorf("URL should contain %s", param)
		}
	}
}

func TestBuildAuthURLCustomScopes(t *testing.T) {
	config := OAuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Scopes:       "read:profile read:sleep",
	}

	url := buildAuthURL(config, "state")

	if !strings.Contains(url, "read%3Aprofile") && !strings.Contains(url, "read:profile") {
		t.Error("URL should contain custom scopes")
	}
}

func TestBuildAuthURLDefaultScopes(t *testing.T) {
	config := OAuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}

	url := buildAuthURL(config, "state")

	// Should contain default scopes including offline for refresh tokens
	if !strings.Contains(url, "offline") {
		t.Error("URL should contain offline scope for refresh tokens")
	}
}

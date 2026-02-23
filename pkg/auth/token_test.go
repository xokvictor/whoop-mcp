package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenIsExpired(t *testing.T) {
	tests := []struct {
		name     string
		expiry   time.Time
		expected bool
	}{
		{
			name:     "zero expiry",
			expiry:   time.Time{},
			expected: false,
		},
		{
			name:     "expired token",
			expiry:   time.Now().Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "expires soon (within 5 min buffer)",
			expiry:   time.Now().Add(3 * time.Minute),
			expected: true,
		},
		{
			name:     "valid token",
			expiry:   time.Now().Add(1 * time.Hour),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{Expiry: tt.expiry}
			if got := token.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTokenExpiresIn(t *testing.T) {
	t.Run("zero expiry", func(t *testing.T) {
		token := &Token{Expiry: time.Time{}}
		if got := token.ExpiresIn(); got != 0 {
			t.Errorf("ExpiresIn() = %v, want 0", got)
		}
	})

	t.Run("future expiry", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour)
		token := &Token{Expiry: expiry}
		got := token.ExpiresIn()
		// Allow some tolerance for test execution time
		if got < 59*time.Minute || got > 61*time.Minute {
			t.Errorf("ExpiresIn() = %v, expected ~1 hour", got)
		}
	})
}

func TestTokenManagerLoadSave(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	tm := &TokenManager{
		tokenPath:    tokenPath,
		clientID:     "test-client",
		clientSecret: "test-secret",
	}

	// Test Load when file doesn't exist
	token, err := tm.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil for non-existent file", err)
	}
	if token != nil {
		t.Error("Load() should return nil for non-existent file")
	}

	// Test Save
	testToken := &Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	if saveErr := tm.Save(testToken); saveErr != nil {
		t.Fatalf("Save() error = %v", saveErr)
	}

	// Verify file permissions
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %v, want 0600", info.Mode().Perm())
	}

	// Test Load
	loaded, err := tm.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.AccessToken != testToken.AccessToken {
		t.Errorf("AccessToken = %v, want %v", loaded.AccessToken, testToken.AccessToken)
	}
	if loaded.RefreshToken != testToken.RefreshToken {
		t.Errorf("RefreshToken = %v, want %v", loaded.RefreshToken, testToken.RefreshToken)
	}
}

func TestTokenManagerDelete(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	tm := &TokenManager{
		tokenPath: tokenPath,
	}

	// Delete non-existent file should not error
	if err := tm.Delete(); err != nil {
		t.Errorf("Delete() non-existent file error = %v", err)
	}

	// Create file and delete
	if err := os.WriteFile(tokenPath, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := tm.Delete(); err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestTokenManagerTokenPath(t *testing.T) {
	tm := &TokenManager{tokenPath: "/test/path/token.json"}
	if got := tm.TokenPath(); got != "/test/path/token.json" {
		t.Errorf("TokenPath() = %v, want /test/path/token.json", got)
	}
}

func TestNewTokenManager(t *testing.T) {
	tm, err := NewTokenManager("client-id", "client-secret")
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	if tm.clientID != "client-id" {
		t.Errorf("clientID = %v, want client-id", tm.clientID)
	}
	if tm.clientSecret != "client-secret" {
		t.Errorf("clientSecret = %v, want client-secret", tm.clientSecret)
	}

	// Token path should end with .whoop/token.json
	if !filepath.IsAbs(tm.tokenPath) {
		t.Error("tokenPath should be absolute")
	}
	if filepath.Base(tm.tokenPath) != "token.json" {
		t.Errorf("tokenPath should end with token.json, got %v", tm.tokenPath)
	}
}

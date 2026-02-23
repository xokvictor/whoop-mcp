package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xokvictor/whoop-mcp/pkg/whoop"
)

const (
	tokenFileName = "token.json"
	dirName       = ".whoop"
	dirPerm       = 0700
	filePerm      = 0600
)

// Token represents OAuth2 token data stored on disk.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

// IsExpired returns true if the token has expired or will expire within 5 minutes.
func (t *Token) IsExpired() bool {
	if t.Expiry.IsZero() {
		return false
	}
	return time.Now().Add(5 * time.Minute).After(t.Expiry)
}

// ExpiresIn returns the duration until the token expires.
func (t *Token) ExpiresIn() time.Duration {
	if t.Expiry.IsZero() {
		return 0
	}
	return time.Until(t.Expiry)
}

// TokenManager handles loading, saving, and refreshing OAuth tokens.
type TokenManager struct {
	tokenPath    string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewTokenManager creates a new TokenManager.
func NewTokenManager(clientID, clientSecret string) (*TokenManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	tokenDir := filepath.Join(homeDir, dirName)
	tokenPath := filepath.Join(tokenDir, tokenFileName)

	return &TokenManager{
		tokenPath:    tokenPath,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// TokenPath returns the path to the token file.
func (tm *TokenManager) TokenPath() string {
	return tm.tokenPath
}

// Load reads the token from disk.
func (tm *TokenManager) Load() (*Token, error) {
	data, err := os.ReadFile(tm.tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token file: %w", err)
	}

	return &token, nil
}

// Save writes the token to disk with appropriate permissions.
func (tm *TokenManager) Save(token *Token) error {
	dir := filepath.Dir(tm.tokenPath)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding token: %w", err)
	}

	if err := os.WriteFile(tm.tokenPath, data, filePerm); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// Delete removes the token file.
func (tm *TokenManager) Delete() error {
	err := os.Remove(tm.tokenPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting token file: %w", err)
	}
	return nil
}

// Refresh obtains a new access token using the refresh token.
func (tm *TokenManager) Refresh(ctx context.Context, refreshToken string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", tm.clientID)
	data.Set("client_secret", tm.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, whoop.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status %d", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("parsing refresh response: %w", err)
	}

	token := &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	if err := tm.Save(token); err != nil {
		return nil, fmt.Errorf("saving refreshed token: %w", err)
	}

	return token, nil
}

// EnsureValidToken loads the token and refreshes it if expired.
// Returns the valid access token or an error if not authenticated.
func (tm *TokenManager) EnsureValidToken(ctx context.Context) (string, error) {
	token, err := tm.Load()
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", nil
	}

	if token.IsExpired() {
		if token.RefreshToken == "" {
			return "", fmt.Errorf("token expired and no refresh token available")
		}
		token, err = tm.Refresh(ctx, token.RefreshToken)
		if err != nil {
			return "", fmt.Errorf("refreshing token: %w", err)
		}
	}

	return token.AccessToken, nil
}

// tokenResponse represents the OAuth2 token response from WHOOP API.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

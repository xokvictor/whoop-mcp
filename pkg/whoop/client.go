package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	BaseURL  = "https://api.prod.whoop.com/developer"
	AuthURL  = "https://api.prod.whoop.com/oauth/oauth2/auth"
	TokenURL = "https://api.prod.whoop.com/oauth/oauth2/token"

	defaultTimeout = 30 * time.Second
)

// TokenProvider is an interface for obtaining valid access tokens.
type TokenProvider interface {
	EnsureValidToken(ctx context.Context) (string, error)
}

// Client is the WHOOP API client
type Client struct {
	httpClient    *http.Client
	baseURL       string
	token         string
	tokenProvider TokenProvider
}

// NewClient creates a new WHOOP API client.
// It reads the access token from WHOOP_ACCESS_TOKEN environment variable.
func NewClient() *Client {
	return NewClientWithToken(os.Getenv("WHOOP_ACCESS_TOKEN"))
}

// NewClientWithToken creates a new WHOOP API client with the specified token.
func NewClientWithToken(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    BaseURL,
		token:      token,
	}
}

// NewClientWithTokenProvider creates a new WHOOP API client with a token provider.
// The provider is used to obtain and refresh tokens automatically.
// If envToken is set, it takes priority over the token provider.
func NewClientWithTokenProvider(envToken string, provider TokenProvider) *Client {
	return &Client{
		httpClient:    &http.Client{Timeout: defaultTimeout},
		baseURL:       BaseURL,
		token:         envToken,
		tokenProvider: provider,
	}
}

// HasToken returns true if the client has an access token configured
// either via environment variable or token provider.
func (c *Client) HasToken() bool {
	return c.token != "" || c.tokenProvider != nil
}

// SetTokenProvider sets the token provider for automatic token management.
func (c *Client) SetTokenProvider(provider TokenProvider) {
	c.tokenProvider = provider
}

func (c *Client) doRequest(ctx context.Context, method, path string) ([]byte, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting token: %w", err)
	}

	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return body, nil
}

// getToken returns the access token to use for requests.
// Priority: 1) env var token 2) token provider 3) empty
func (c *Client) getToken(ctx context.Context) (string, error) {
	// Environment variable token has highest priority
	if c.token != "" {
		return c.token, nil
	}

	// Try token provider
	if c.tokenProvider != nil {
		token, err := c.tokenProvider.EnsureValidToken(ctx)
		if err != nil {
			return "", err
		}
		if token != "" {
			return token, nil
		}
	}

	return "", nil
}

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	body, err := c.doRequest(ctx, http.MethodGet, path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	return nil
}

// APIError represents an error response from the WHOOP API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// IsUnauthorized returns true if the error is a 401 Unauthorized response.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsNotFound returns true if the error is a 404 Not Found response.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if the error is a 429 Too Many Requests response.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

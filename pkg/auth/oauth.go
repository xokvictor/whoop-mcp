package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/xokvictor/whoop-mcp/pkg/whoop"
)

const (
	callbackPort  = 8080
	callbackPath  = "/callback"
	authTimeout   = 5 * time.Minute
	defaultScopes = "read:profile read:body_measurement read:cycles read:recovery read:sleep read:workout offline"
	redirectURI   = "http://localhost:8080/callback"
)

// OAuthConfig contains OAuth configuration.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	Scopes       string
}

// AuthResult contains the result of the OAuth authorization flow.
type AuthResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Email   string `json:"email,omitempty"`
}

// StartAuthFlow initiates the OAuth authorization flow.
// It starts a local HTTP server, opens the browser, and waits for the callback.
func StartAuthFlow(ctx context.Context, config OAuthConfig, tokenManager *TokenManager) (*AuthResult, error) {
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server, err := startCallbackServer(state, codeChan, errChan)
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}
	defer server.Shutdown(context.Background())

	authURL := buildAuthURL(config, state)

	if err := openBrowser(authURL); err != nil {
		return &AuthResult{
			Success: false,
			Message: fmt.Sprintf("Failed to open browser. Please visit this URL manually:\n%s", authURL),
		}, nil
	}

	select {
	case code := <-codeChan:
		token, err := exchangeCode(ctx, config, code)
		if err != nil {
			return nil, fmt.Errorf("exchanging code: %w", err)
		}

		if err := tokenManager.Save(token); err != nil {
			return nil, fmt.Errorf("saving token: %w", err)
		}

		return &AuthResult{
			Success: true,
			Message: "Authorization successful! Token saved.",
		}, nil

	case err := <-errChan:
		return nil, err

	case <-time.After(authTimeout):
		return &AuthResult{
			Success: false,
			Message: "Authorization timed out. Please try again.",
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func buildAuthURL(config OAuthConfig, state string) string {
	scopes := config.Scopes
	if scopes == "" {
		scopes = defaultScopes
	}

	params := url.Values{}
	params.Set("client_id", config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", scopes)
	params.Set("state", state)

	return whoop.AuthURL + "?" + params.Encode()
}

func startCallbackServer(expectedState string, codeChan chan<- string, errChan chan<- error) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			errChan <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errDesc := r.URL.Query().Get("error_description")
			errChan <- fmt.Errorf("authorization error: %s - %s", errMsg, errDesc)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><h1>Authorization Failed</h1><p>%s</p><script>setTimeout(function(){window.close();},3000);</script></body></html>`, html.EscapeString(errDesc))
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Authorization Successful!</h1><p>You can close this window.</p><script>setTimeout(function(){window.close();},3000);</script></body></html>`)
		codeChan <- code
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", callbackPort))
	if err != nil {
		return nil, fmt.Errorf("port %d is already in use: %w", callbackPort, err)
	}

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	return server, nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func exchangeCode(ctx context.Context, config OAuthConfig, code string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, whoop.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	return &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

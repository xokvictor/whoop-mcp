package whoop

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.baseURL != BaseURL {
		t.Errorf("expected baseURL %q, got %q", BaseURL, client.baseURL)
	}
}

func TestNewClientWithToken(t *testing.T) {
	token := "test-token"
	client := NewClientWithToken(token)

	if client == nil {
		t.Fatal("NewClientWithToken() returned nil")
	}
	if client.token != token {
		t.Errorf("expected token %q, got %q", token, client.token)
	}
	if !client.HasToken() {
		t.Error("HasToken() should return true when token is set")
	}
}

func TestHasToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"empty token", "", false},
		{"valid token", "some-token", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientWithToken(tt.token)
			if got := client.HasToken(); got != tt.expected {
				t.Errorf("HasToken() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 401, Message: "Unauthorized"}

	if err.Error() != "API error (status 401): Unauthorized" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !err.IsUnauthorized() {
		t.Error("IsUnauthorized() should return true for 401")
	}
	if err.IsNotFound() {
		t.Error("IsNotFound() should return false for 401")
	}
	if err.IsRateLimited() {
		t.Error("IsRateLimited() should return false for 401")
	}
}

func TestAPIErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		isUnauthorized bool
		isNotFound     bool
		isRateLimited  bool
	}{
		{"401 Unauthorized", 401, true, false, false},
		{"404 Not Found", 404, false, true, false},
		{"429 Too Many Requests", 429, false, false, true},
		{"500 Internal Server Error", 500, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode, Message: "test"}

			if got := err.IsUnauthorized(); got != tt.isUnauthorized {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.isUnauthorized)
			}
			if got := err.IsNotFound(); got != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.isNotFound)
			}
			if got := err.IsRateLimited(); got != tt.isRateLimited {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.isRateLimited)
			}
		})
	}
}

func TestClientDoRequest(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Error("missing or incorrect Authorization header")
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		client := NewClientWithToken("test-token")
		client.baseURL = server.URL

		body, err := client.doRequest(context.Background(), http.MethodGet, "/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := `{"status": "ok"}`
		if string(body) != expected {
			t.Errorf("expected body %q, got %q", expected, string(body))
		}
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
		}))
		defer server.Close()

		client := NewClientWithToken("bad-token")
		client.baseURL = server.URL

		_, err := client.doRequest(context.Background(), http.MethodGet, "/test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("expected *APIError, got %T", err)
		}
		if apiErr.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", apiErr.StatusCode)
		}
	})

	t.Run("no token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				t.Error("Authorization header should be empty")
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := NewClientWithToken("")
		client.baseURL = server.URL

		_, err := client.doRequest(context.Background(), http.MethodGet, "/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"name": "test"})
	}))
	defer server.Close()

	client := NewClientWithToken("test-token")
	client.baseURL = server.URL

	var result struct {
		Name string `json:"name"`
	}

	err := client.get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("expected name 'test', got %q", result.Name)
	}
}

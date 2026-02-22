package whoop

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid UUID lowercase", "123e4567-e89b-12d3-a456-426614174000", true},
		{"valid UUID uppercase", "123E4567-E89B-12D3-A456-426614174000", true},
		{"valid UUID mixed case", "123e4567-E89B-12d3-A456-426614174000", true},
		{"empty string", "", false},
		{"invalid format", "not-a-uuid", false},
		{"missing dashes", "123e4567e89b12d3a456426614174000", false},
		{"too short", "123e4567-e89b-12d3-a456", false},
		{"too long", "123e4567-e89b-12d3-a456-4266141740001234", false},
		{"invalid characters", "123g4567-e89b-12d3-a456-426614174000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidUUID(tt.input); got != tt.expected {
				t.Errorf("isValidUUID(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name        string
		start       string
		end         string
		limit       int
		nextToken   string
		contains    []string
		notContains []string
	}{
		{
			name:     "empty params",
			contains: []string{},
		},
		{
			name:     "start only",
			start:    "2024-01-01",
			contains: []string{"start=2024-01-01"},
		},
		{
			name:      "all params",
			start:     "2024-01-01",
			end:       "2024-12-31",
			limit:     10,
			nextToken: "abc123",
			contains:  []string{"start=2024-01-01", "end=2024-12-31", "limit=10", "nextToken=abc123"},
		},
		{
			name:        "limit over max",
			limit:       100,
			contains:    []string{"limit=25"},
			notContains: []string{"limit=100"},
		},
		{
			name:        "zero limit",
			limit:       0,
			notContains: []string{"limit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildQuery(tt.start, tt.end, tt.limit, tt.nextToken)

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected query to contain %q, got %q", s, result)
				}
			}

			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("expected query NOT to contain %q, got %q", s, result)
				}
			}
		})
	}
}

func TestGetUserProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/user/profile/basic" {
			t.Errorf("expected path /v2/user/profile/basic, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserBasicProfile{
			UserID:    123,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		})
	}))
	defer server.Close()

	client := NewClientWithToken("test-token")
	client.baseURL = server.URL

	profile, err := client.GetUserProfile(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.UserID != 123 {
		t.Errorf("expected UserID 123, got %d", profile.UserID)
	}
	if profile.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", profile.Email)
	}
}

func TestGetCycleByID(t *testing.T) {
	t.Run("valid cycle ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/cycle/123" {
				t.Errorf("expected path /v2/cycle/123, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Cycle{ID: 123})
		}))
		defer server.Close()

		client := NewClientWithToken("test-token")
		client.baseURL = server.URL

		cycle, err := client.GetCycleByID(context.Background(), 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cycle.ID != 123 {
			t.Errorf("expected ID 123, got %d", cycle.ID)
		}
	})

	t.Run("invalid cycle ID", func(t *testing.T) {
		client := NewClientWithToken("test-token")

		_, err := client.GetCycleByID(context.Background(), 0)
		if err == nil {
			t.Fatal("expected error for invalid cycle ID")
		}
		if !strings.Contains(err.Error(), "invalid cycle ID") {
			t.Errorf("expected 'invalid cycle ID' error, got %v", err)
		}

		_, err = client.GetCycleByID(context.Background(), -1)
		if err == nil {
			t.Fatal("expected error for negative cycle ID")
		}
	})
}

func TestGetSleepByID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		sleepID := "123e4567-e89b-12d3-a456-426614174000"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/v2/activity/sleep/" + sleepID
			if r.URL.Path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Sleep{ID: sleepID})
		}))
		defer server.Close()

		client := NewClientWithToken("test-token")
		client.baseURL = server.URL

		sleep, err := client.GetSleepByID(context.Background(), sleepID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sleep.ID != sleepID {
			t.Errorf("expected ID %s, got %s", sleepID, sleep.ID)
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		client := NewClientWithToken("test-token")

		_, err := client.GetSleepByID(context.Background(), "not-a-uuid")
		if err == nil {
			t.Fatal("expected error for invalid UUID")
		}
		if !strings.Contains(err.Error(), "invalid sleep ID") {
			t.Errorf("expected 'invalid sleep ID' error, got %v", err)
		}
	})
}

func TestGetWorkoutByID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		workoutID := "123e4567-e89b-12d3-a456-426614174000"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(WorkoutV2{ID: workoutID})
		}))
		defer server.Close()

		client := NewClientWithToken("test-token")
		client.baseURL = server.URL

		workout, err := client.GetWorkoutByID(context.Background(), workoutID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if workout.ID != workoutID {
			t.Errorf("expected ID %s, got %s", workoutID, workout.ID)
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		client := NewClientWithToken("test-token")

		_, err := client.GetWorkoutByID(context.Background(), "invalid")
		if err == nil {
			t.Fatal("expected error for invalid UUID")
		}
	})
}

func TestGetCycles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/cycle" {
			t.Errorf("expected path /v2/cycle, got %s", r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PaginatedCycleResponse{
			Records: []Cycle{{ID: 1}, {ID: 2}},
		})
	}))
	defer server.Close()

	client := NewClientWithToken("test-token")
	client.baseURL = server.URL

	result, err := client.GetCycles(context.Background(), CycleParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Records) != 2 {
		t.Errorf("expected 2 records, got %d", len(result.Records))
	}
}

func TestGetActivityMapping(t *testing.T) {
	t.Run("valid activity ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/activity-mapping/123" {
				t.Errorf("expected path /v1/activity-mapping/123, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ActivityIdMappingResponse{V2ActivityID: "uuid-123"})
		}))
		defer server.Close()

		client := NewClientWithToken("test-token")
		client.baseURL = server.URL

		mapping, err := client.GetActivityMapping(context.Background(), 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mapping.V2ActivityID != "uuid-123" {
			t.Errorf("expected V2ActivityID uuid-123, got %s", mapping.V2ActivityID)
		}
	})

	t.Run("invalid activity ID", func(t *testing.T) {
		client := NewClientWithToken("test-token")

		_, err := client.GetActivityMapping(context.Background(), 0)
		if err == nil {
			t.Fatal("expected error for invalid activity ID")
		}
	})
}

package main

import (
	"testing"
)

func TestGetStringArg(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "existing string key",
			args:     map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "missing key",
			args:     map[string]interface{}{"other": "value"},
			key:      "key",
			expected: "",
		},
		{
			name:     "nil args",
			args:     nil,
			key:      "key",
			expected: "",
		},
		{
			name:     "non-string value",
			args:     map[string]interface{}{"key": 123},
			key:      "key",
			expected: "",
		},
		{
			name:     "empty string value",
			args:     map[string]interface{}{"key": ""},
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringArg(tt.args, tt.key)
			if result != tt.expected {
				t.Errorf("getStringArg() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetIntArg(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal int
		expected   int
	}{
		{
			name:       "existing float64 key",
			args:       map[string]interface{}{"key": float64(42)},
			key:        "key",
			defaultVal: 0,
			expected:   42,
		},
		{
			name:       "existing int key",
			args:       map[string]interface{}{"key": 42},
			key:        "key",
			defaultVal: 0,
			expected:   42,
		},
		{
			name:       "existing int64 key",
			args:       map[string]interface{}{"key": int64(42)},
			key:        "key",
			defaultVal: 0,
			expected:   42,
		},
		{
			name:       "missing key returns default",
			args:       map[string]interface{}{"other": 42},
			key:        "key",
			defaultVal: 10,
			expected:   10,
		},
		{
			name:       "nil args returns default",
			args:       nil,
			key:        "key",
			defaultVal: 5,
			expected:   5,
		},
		{
			name:       "non-numeric value returns default",
			args:       map[string]interface{}{"key": "not a number"},
			key:        "key",
			defaultVal: 7,
			expected:   7,
		},
		{
			name:       "zero value",
			args:       map[string]interface{}{"key": float64(0)},
			key:        "key",
			defaultVal: 10,
			expected:   0,
		},
		{
			name:       "negative value",
			args:       map[string]interface{}{"key": float64(-5)},
			key:        "key",
			defaultVal: 10,
			expected:   -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntArg(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("getIntArg() = %d, want %d", result, tt.expected)
			}
		})
	}
}
